package lbs

import (
	"errors"
	"log"
	"math"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"
)

var (
	ErrEmptyRequest = errors.New("lbs: empty request")
	ErrNotFound     = errors.New("lbs: not found")
)

// CollectionName описывает название коллекции с данными для LBS.
var CollectionName = "lbs"

// DB описывает хранилище LBS данных и работу с ними.
type DB struct {
	*mongo.DB // хранилище
}

// InitDB возвращает инициализированный объект для работы с хранилищем LBS данных.
func InitDB(mdb *mongo.DB) (db *DB, err error) {
	db = &DB{mdb}
	coll := mdb.GetCollection(CollectionName)
	err = coll.EnsureIndex(mgo.Index{
		Key:      []string{"mcc", "mnc", "area", "cell"},
		Unique:   true,
		DropDups: true,
	})
	mdb.FreeCollection(coll)
	return
}

// Key описывает ключ для поиска информации по LBS.
type Key struct {
	MCC  uint16 // country code  (250 - Россия, 255 - Украина, Беларусь - 257)
	MNC  uint32 // operator code
	Area uint16 // the base station cell number
	Cell uint32 // base station number
}

// Search ищет и вычисляет координаты, переданные в запросе, на основании данных вышек сотовой
// связи. Если данных не достаточно или необходимая для вычислений информация не найдена в
// хранилище, то возвращается ошибка.
func (db *DB) Search(req *Request) (point *geo.Point, err error) {
	if req == nil {
		return nil, ErrEmptyRequest
	}
	coll := db.GetCollection(CollectionName)
	defer db.FreeCollection(coll)
	search := Key{
		MCC: req.MCC,
		MNC: req.MNC,
	}
	selector := bson.M{"point": 1, "_id": 0}
	var sm, slat, slon float64
	for _, cell := range req.Cells {
		search.Area = cell.Area
		search.Cell = cell.ID
		var data struct{ geo.Point }
		err := coll.Find(search).Select(selector).One(&data)
		switch err {
		case mgo.ErrNotFound: // не найдено в базе - игнорируем
			log.Println("LBS Search:", search, "-", "not found")
			continue
		case nil: // найдено - продолжаем
			log.Println("LBS Search:", search, "-", data.Point)
			m := math.Pow(10, (float64(cell.DBM)/20)) * 1000
			sm += m
			slon += data.Point.Longitude() * m
			slat += data.Point.Latitude() * m
		default: // ошибка работы с хранилищем
			return nil, err
		}
	}
	if sm == 0 {
		return nil, ErrNotFound
	}
	point = geo.NewPoint(slon/sm, slat/sm)
	log.Println("LBS Result:", point)
	return
}

// SearchLBS ищет координаты, разбирая предварительно строку с запросом в формате LBS.
func (db *DB) SearchLBS(s string) (point *geo.Point, err error) {
	req, err := Parse(s)
	if err != nil {
		return
	}
	return db.Search(req)
}
