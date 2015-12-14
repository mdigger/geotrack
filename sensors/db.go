package sensors

import (
	"time"

	"github.com/mdigger/geotrack/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var CollectionName = "sensors"

var ExpireAfter = time.Duration(time.Hour * 24 * 31)

type DB struct {
	*mongo.DB // соединение с MongoDB
}

func InitDB(mdb *mongo.DB) (db *DB, err error) {
	db = &DB{mdb}
	coll := mdb.GetCollection(CollectionName)
	defer mdb.FreeCollection(coll)
	if err = coll.EnsureIndex(mgo.Index{
		Key:         []string{"time"},
		ExpireAfter: ExpireAfter,
	}); err != nil {
		return
	}
	if err = coll.EnsureIndexKey("groupid", "deviceid", "-_id"); err != nil {
		return
	}
	return
}

// SensorData описывает входящий формат данных с сенсорами.
type SensorData struct {
	GroupID  string                 // идентификатор группы
	DeviceID string                 // уникальный идентификатор устройства
	Time     time.Time              // временная метка
	Data     map[string]interface{} // именованные значения датчиков
}

// Add добавляет записи сенсоров в хранилище.
func (db *DB) Add(sensors ...SensorData) (err error) {
	// конвертируем типизированный список в нетипизированный
	data := make([]interface{}, len(sensors))
	for i, item := range sensors {
		data[i] = item
	}
	// сохраняем в коллекции
	coll := db.GetCollection(CollectionName)
	err = coll.Insert(data...)
	db.FreeCollection(coll)
	return
}

// Track описывает единичные данные трекинга, содержащие координаты точки, время их получения
// и уникальный идентификатор данных в хранилище. Идентификатор не является обязательным:
// если он отсутствует, то будет автоматически присвоен при сохранении в хранилище.
type Sensor struct {
	ID   bson.ObjectId          `bson:"_id"` // уникальный идентификатор
	Time time.Time              // временная метка
	Data map[string]interface{} // именованные значения датчиков
}

// selector описывает список выбираемых полей
var selector = bson.M{"groupid": 0, "deviceid": 0}

// Get возвращает список треков для указанного устройства.
//
// Метод поддерживает разбиение результатов на отдельные блоки: limit указывает максимальное
// количество отдаваемых в ответ данных, а lastID — идентификатор последнего полученного трека.
func (db *DB) Get(groupID, deviceID string, limit int, lastID string) (sensors []Sensor, err error) {
	coll := db.GetCollection(CollectionName)
	// ищем все треки с указанного устройства
	var search = bson.M{
		"groupid":  groupID,
		"deviceid": deviceID,
	}
	if lastID != "" && bson.IsObjectIdHex(lastID) {
		search["_id"] = bson.M{"$lt": bson.ObjectIdHex(lastID)} // старее последнего полученного идентификатора
	}
	// используем обратную сортировку: свежие записи должны идти раньше более старых
	query := coll.Find(search).Select(selector).Sort("-$natural")
	if limit > 0 {
		query.Limit(limit)                 // ограничиваем количество запрашиваемых данных
		sensors = make([]Sensor, 0, limit) // мы заранее знаем максимальное количество записей
	} else {
		sensors = make([]Sensor, 0)
	}
	err = query.All(&sensors)
	db.FreeCollection(coll)
	return
}

// GetDay возвращает список треков для указанного устройства за последние сутки.
func (db *DB) GetDay(groupID, deviceID string) (sensors []Sensor, err error) {
	coll := db.GetCollection(CollectionName)
	// ищем все треки с указанного устройства за последние сутки
	var search = bson.M{
		"groupid":  groupID,
		"deviceid": deviceID,
		"time":     bson.M{"$gt": time.Now().Add(-24 * time.Hour)},
	}
	// используем обратную сортировку: свежие записи должны идти раньше более старых
	query := coll.Find(search).Select(selector).Sort("-$natural")
	sensors = make([]Sensor, 0)
	err = query.All(&sensors)
	db.FreeCollection(coll)
	return
}
