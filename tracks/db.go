package tracks

import (
	"time"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	// CollectionName описывает название коллекции с данными.
	CollectionName = "tracks"

	// ExpireAfter описывает время жизни записи трека в хранилище, после которого
	// эта запись автоматически удаляется. По умолчанию время жизни записи задана как одна неделя.
	//
	// Необходимо обратить внимание, что значение этой переменной используется при создании индекса
	// удаления данных на MongoDB, поэтому ее изменение может повлиять только при пересоздании
	// этого индекса или при инициализации новой базы данных.
	ExpireAfter = time.Duration(time.Hour * 24 * 7)
)

// DB описывает интерфейс для работы с хранилищем данных трекинга для устройств.
type DB struct {
	*mongo.DB // соединение с MongoDB
}

// InitDB инициализирует работу с хранилищем данных трекинга и возвращает объект для работы
// с ними. В процесс инициализации проверяется наличие всех необходимых индексов для удобной
// работы с данными и если их нет, то они автоматически создаются, что позволяет использовать
// данный класс даже для работы с пустой базовй данных.
func InitDB(mdb *mongo.DB) (db *DB, err error) {
	db = &DB{mdb}
	coll := mdb.GetCollection(CollectionName)
	defer mdb.FreeCollection(coll)
	// ключ для выборки треков по идентификатору устройства
	// используется для получения списка треков для конкретного устройства
	if err = coll.EnsureIndexKey("groupid", "deviceid", "-_id"); err != nil {
		return
	}
	// ключ для выборки треков по идентификатору устройства и времени
	if err = coll.EnsureIndexKey("groupid", "deviceid", "time", "-_id"); err != nil {
		return
	}
	// ключ для автоматического удаления устаревших записей трекинга устройств
	if err = coll.EnsureIndex(mgo.Index{
		Key:         []string{"time"},
		ExpireAfter: ExpireAfter,
	}); err != nil {
		return
	}
	return
}

// TrackData описывает входящий формат данных трекинга.
type TrackData struct {
	GroupID  string     // идентификатор группы
	DeviceID string     // уникальный идентификатор устройства
	Time     time.Time  // временная метка
	Point    *geo.Point // координаты точки
}

// Add добавляет запись трекинга в хранилище.
func (db *DB) Add(track *TrackData) (err error) {
	coll := db.GetCollection(CollectionName)
	err = coll.Insert(track)
	db.FreeCollection(coll)
	return
}

// Track описывает единичные данные трекинга, содержащие координаты точки, время их получения
// и уникальный идентификатор данных в хранилище. Идентификатор не является обязательным:
// если он отсутствует, то будет автоматически присвоен при сохранении в хранилище.
type Track struct {
	ID    bson.ObjectId `bson:"_id"` // уникальный идентификатор
	Time  time.Time     // временная метка
	Point geo.Point     // координаты точки
}

// selector описывает список выбираемых полей
var selector = bson.M{"time": 1, "point": 1}

// Get возвращает список треков для указанного устройства.
//
// Метод поддерживает разбиение результатов на отдельные блоки: limit указывает максимальное
// количество отдаваемых в ответ данных, а lastID — идентификатор последнего полученного трека.
func (db *DB) Get(groupID, deviceID string, limit int, lastID string) (tracks []*Track, err error) {
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
		query.Limit(limit)                // ограничиваем количество запрашиваемых данных
		tracks = make([]*Track, 0, limit) // мы заранее знаем максимальное количество записей
	} else {
		tracks = make([]*Track, 0)
	}
	err = query.All(&tracks)
	db.FreeCollection(coll)
	return
}

// GetDay возвращает список треков для указанного устройства за последние сутки.
func (db *DB) GetDay(groupID, deviceID string) (tracks []*Track, err error) {
	coll := db.GetCollection(CollectionName)
	// ищем все треки с указанного устройства за последние сутки
	var search = bson.M{
		"groupid":  groupID,
		"deviceid": deviceID,
		"time":     bson.M{"$gt": time.Now().Add(-24 * time.Hour)},
	}
	// используем обратную сортировку: свежие записи должны идти раньше более старых
	query := coll.Find(search).Select(selector).Sort("-$natural")
	tracks = make([]*Track, 0)
	err = query.All(&tracks)
	db.FreeCollection(coll)
	return
}

// GetLast возвращает самый последний трек для данного устройства, сохраненный
// в хранилище.
func (db *DB) GetLast(groupID, deviceID string) (track *Track, err error) {
	coll := db.GetCollection(CollectionName)
	track = new(Track)
	var search = bson.M{
		"groupid":  groupID,
		"deviceid": deviceID,
	}
	err = coll.Find(search).Select(selector).Sort("-$natural").One(track)
	db.FreeCollection(coll)
	return
}

// GetDevicesID возвращает список всех идентификаторов устройства, найденных в хранилище
// с данными трекинга для данной группы.
func (db *DB) GetDevicesID(groupID string) (deviceids []string, err error) {
	coll := db.GetCollection(CollectionName)
	deviceids = make([]string, 0)
	err = coll.Find(bson.M{"groupid": groupID}).Distinct("deviceid", &deviceids)
	db.FreeCollection(coll)
	return
}
