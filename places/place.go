package places

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
// В объекте должно быть указано хотя бы одно описание места: либо круг, либо полигон.
// Если указано и то, и другое, то используется только круг. Если не указано ни того,
// ни другого, то такая запись либо игнорируется, либо удаляется, если указан корректный
// ее идентификатор. В процессе сохранения всегда идентификатор группы заменяется на
// указанный в параметрах вызова.
func (db *DB) Save(groupID string, places ...*Place) (err error) {
	coll := db.GetCollection(CollectionName)
	for _, place := range places {
		// анализируем описание места и формируем данные для индексации
		if place.Circle != nil {
			place.Polygon = nil
			place.Geo = place.Circle.Geo()
		} else if place.Polygon != nil {
			place.Circle = nil
			place.Geo = place.Polygon.Geo()
		} else {
			// удаляем, если нет данных и нормальный идентификатор
			// в противном случае — просто игнорируем
			if place.ID.Valid() {
				if err = coll.RemoveId(place.ID); err != nil {
					break
				}
			}
			continue
		}
		place.GroupID = groupID // восстанавливаем группу, если вдруг она пропала
		if !place.ID.Valid() {
			place.ID = bson.NewObjectId()
		}
		if _, err = coll.UpsertId(place.ID, place); err != nil {
			break
		}
	}
	db.FreeCollection(coll)
	return
}

// Get возвращает список всех описаний мест для указанной группы.
// Результат содержит только информацию с описание круга или полигона. Информация
// о группе и сформированном внутреннем индексном объекте Geo не возвращается.
func (db *DB) Get(groupID string) (places []*Place, err error) {
	coll := db.GetCollection(CollectionName)
	places = make([]*Place, 0)
	selector := bson.M{"groupid": 0, "geo": 0}
	err = coll.Find(bson.M{"groupid": groupID}).Select(selector).All(&places)
	db.FreeCollection(coll)
	return
}

// Track возвращает список всех идентификаторов мест, определенных для группы,
// которым соответствует данная точка трекера.
func (db *DB) Track(track *tracks.TrackData) (placeIDs []string, err error) {
	coll := db.GetCollection(CollectionName)
	ids := make([]*bson.ObjectId, 0)
	err = coll.Find(bson.M{
		"groupid": track.GroupID,
		"geo": bson.M{
			"$geoIntersects": bson.M{
				"$geometry": track.Point.Geo(),
			},
		},
	}).Distinct("_id", &ids)
	placeIDs = make([]string, len(ids))
	for i, id := range ids {
		placeIDs[i] = id.Hex()
	}
	return
}
