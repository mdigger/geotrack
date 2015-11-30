package ublox

import (
	"log"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"
)

var (
	CollectionName = "ublox"                         // название коллекции
	ExpireAfter    = time.Duration(time.Minute * 30) // время жизни элемента кеша
	MaxDistance    = 100000                          // дистанция в метрах для выборки
)

// Cache описывает кеш ответов сервера U-blox с эфемеридами.
type Cache struct {
	*mongo.DB         // соединение с MongoDB
	client    *Client // клиент для доступа к информации U-blox
}

// InitCache возвращает инициализированное хранилище кеша для данных с эфемеридами.
// В процессе инициализации проверяет наличие необходимых индексов и создает, в случе их
// отсутствия. Если идексы уже существуют, но отличиются от тех, что задаются по умолчанию,
// то возвращает ошибку.
func InitCache(mdb *mongo.DB, token string) (cache *Cache, err error) {
	cache = &Cache{
		DB:     mdb,
		client: NewClient(token),
	}
	coll := mdb.GetCollection(CollectionName)
	defer mdb.FreeCollection(coll)
	if err = coll.EnsureIndexKey("profile", "$2dsphere:point"); err != nil {
		return
	}
	if err = coll.EnsureIndex(mgo.Index{
		Key:         []string{"time"},
		ExpireAfter: ExpireAfter,
	}); err != nil {
		return
	}
	return
}

// storeData описывает формат данных для хранения.
type storeData struct {
	*Profile                // профиль
	Point    *geo.GeoObject // координаты
	Data     []byte         // содержимое ответа
	Time     time.Time      // временная метка
}

// Get возвращает данные эфемерид для указанной точки. Данные возвращаются из кеша, если
// есть для близлежайшей точки, либо запрашиваются с сервера U-blox в противном случае.
func (c *Cache) Get(point *geo.Point, profile *Profile) (data []byte, err error) {
	coll := c.GetCollection(CollectionName)
	defer c.FreeCollection(coll)
	// сначала ищем в кеше
	var cacheData struct {
		Data []byte
	}
	err = coll.Find(bson.M{
		"profile": profile,
		"point": bson.M{
			"$near":        point.Geo(),
			"$maxDistance": MaxDistance,
		}}).Select(bson.M{"data": 1, "_id": 0}).One(&cacheData)
	// if err != mgo.ErrNotFound {
	switch err {
	case nil: // данные получены из кеша
		data = cacheData.Data
		return
	case mgo.ErrNotFound: // данные в кеше не найдены - запрашиваем у сервера
	default: // ошибка получения данных
		log.Println("UBLOX cache error:", err)
	}
	// в кеше ничего не нашли... нужно запрашивать.
	data, err = c.client.GetOnline(point, profile)
	if err != nil {
		return
	}
	// сохраняем ответ в хранилище
	err = coll.Insert(&storeData{
		Profile: profile,
		Point:   point.Geo(),
		Data:    data,
		Time:    time.Now(),
	})
	return
}
