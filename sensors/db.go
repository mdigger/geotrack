package sensors

import (
	"time"

	"github.com/kr/pretty"
	"github.com/mdigger/geotrack/mongo"
)

var CollectionName = "sensors"

var ExpireAfter = time.Duration(time.Hour * 24 * 7)

type DB struct {
	*mongo.DB // соединение с MongoDB
}

func InitDB(mdb *mongo.DB) (db *DB, err error) {
	db = &DB{mdb}
	coll := mdb.GetCollection(CollectionName)
	defer mdb.FreeCollection(coll)
	// ключ для выборки треков по идентификатору устройства и времени
	if err = coll.EnsureIndexKey("groupid", "deviceid", "time", "-_id"); err != nil {
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
	if err != nil {
		pretty.Println(err)
	}
	return
}