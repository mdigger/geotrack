package places

import (
	"errors"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/tracks"
	"gopkg.in/mgo.v2/bson"
)

var CollectionName = "places"

type DB struct {
	*mongo.DB // соединение с MongoDB
}

func InitDB(mdb *mongo.DB) (db *DB, err error) {
	db = &DB{mdb}
	coll := mdb.GetCollection(CollectionName)
	defer mdb.FreeCollection(coll)
	if err = coll.EnsureIndexKey("groupid", "$2dsphere:geo"); err != nil {
		return
	}
	return
}

// Place описывает географическое место.
type Place struct {
	ID      bson.ObjectId `bson:"_id"`                          // уникальный идентификатор
	Name    string        `json:",omitempty"`                   // название
	GroupID string        `json:",omitempty"`                   // идентификатор группы
	Circle  *geo.Circle   `json:",omitempty" bson:",omitempty"` // описывает круг
	Polygon *geo.Polygon  `json:",omitempty" bson:",omitempty"` // описывает область
	Geo     interface{}   `json:"-"`                            // описание координат места для поиска
}

var (
	// ErrBadPlaceData возвращается, если в описании места не указано ни круга, ни полигона.
	ErrBadPlaceData = errors.New("bad place data")
	ErrBadID        = errors.New("bad place id")
)

func (db *DB) Get(groupID string, placeID bson.ObjectId) (place *Place, err error) {
	coll := db.GetCollection(CollectionName)
	place = new(Place)
	selector := bson.M{"groupid": 0, "geo": 0}
	err = coll.Find(bson.M{"_id": placeID, "groupid": groupID}).Select(selector).All(&place)
	db.FreeCollection(coll)
	return
}

// GetAll возвращает список всех описаний мест для указанной группы.
// Результат содержит только информацию с описание круга или полигона. Информация
// о группе и сформированном внутреннем индексном объекте Geo не возвращается.
func (db *DB) GetAll(groupID string) (places []Place, err error) {
	coll := db.GetCollection(CollectionName)
	places = make([]Place, 0)
	selector := bson.M{"groupid": 0, "geo": 0}
	err = coll.Find(bson.M{"groupid": groupID}).Select(selector).All(&places)
	db.FreeCollection(coll)
	return
}

// Save сохраняет описание места в хранилище.
// В объекте должно быть указано хотя бы одно описание места: либо круг, либо полигон.
// Если указано и то, и другое, то используется только круг. Если не указано ни того,
// ни другого, то такая запись игнорируется.
func (db *DB) Save(place Place) (id bson.ObjectId, err error) {
	coll := db.GetCollection(CollectionName)
	defer db.FreeCollection(coll)
	// анализируем описание места и формируем данные для индексации
	if place.Circle != nil {
		place.Polygon = nil
		place.Geo = place.Circle.Geo()
	} else if place.Polygon != nil {
		place.Circle = nil
		place.Geo = place.Polygon.Geo()
	} else {
		return "", ErrBadPlaceData
	}
	if !place.ID.Valid() {
		place.ID = bson.NewObjectId()
	}
	if _, err = coll.UpsertId(place.ID, place); err != nil {
		return "", err
	}
	return place.ID, nil
}

func (db *DB) Delete(groupID string, placeID bson.ObjectId) (err error) {
	coll := db.GetCollection(CollectionName)
	err = coll.Remove(bson.M{"_id": placeID, "groupid": groupID})
	db.FreeCollection(coll)
	return
}

// Track возвращает список всех идентификаторов мест, определенных для группы,
// которым соответствует данная точка трекера.
func (db *DB) Track(track tracks.TrackData) (placeIDs []bson.ObjectId, err error) {
	coll := db.GetCollection(CollectionName)
	placeIDs = make([]bson.ObjectId, 0)
	err = coll.Find(bson.M{
		"groupid": track.GroupID,
		"geo": bson.M{
			"$geoIntersects": bson.M{
				"$geometry": track.Location.Geo(),
			},
		},
	}).Distinct("_id", &placeIDs)
	return
}
