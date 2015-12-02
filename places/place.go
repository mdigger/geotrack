package place

import (
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

// Save сохраняет описание места или нескольких для указанной группы в хранилище.
func (db *DB) Save(groupID string, places ...*Place) (err error) {
	coll := db.GetCollection(CollectionName)
	for _, place := range places {
		if !place.ID.Valid() {
			place.ID = bson.NewObjectId()
		}
		place.GroupID = groupID // восстанавливаем группу, если вдруг она пропала
		// анализируем описание места и формируем данные для индексации
		if place.Circle != nil {
			place.Polygon = nil
			place.Geo = place.Circle.Geo()
		} else if place.Polygon != nil {
			place.Circle = nil
			place.Geo = place.Polygon.Geo()
		} else {
			continue
		}
		if _, err = coll.UpsertId(place.ID, place); err != nil {
			break
		}
	}
	db.FreeCollection(coll)
	return
}

// Get возвращает список всех описаний мест для указанной группы.
func (db *DB) Get(groupID string) (places []*Place, err error) {
	coll := db.GetCollection(CollectionName)
	places = make([]*Place, 0)
	selector := bson.M{"groupid": 0, "geo": 0}
	err = coll.Find(bson.M{"groupid": groupID}).Select(selector).All(&places)
	db.FreeCollection(coll)
	return
}

// Track возвращает список всех идентификаторов мест, которым соответствует данная точка трекера.
func (db *DB) Track(track *tracks.TrackData) (placeIDs []string, err error) {
	coll := db.GetCollection(CollectionName)
	placeIDs = make([]string, 0)
	err = coll.Find(bson.M{
		"groupid": track.GroupID,
		"geo": bson.M{
			"$geoIntersects": bson.M{
				"$geometry": track.Point.Geo(),
			},
		},
	}).Distinct("_id", &placeIDs)
	return
}
