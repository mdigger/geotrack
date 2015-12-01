package users

import (
	"fmt"

	"github.com/mdigger/geotrack/mongo"
	"gopkg.in/mgo.v2/bson"
)

var CollectionName = "users" // содержит название коллекции с данными треков

// DB описывает интерфейс для работы с хранилищем данных с информацие о зарегистрированных
// в системе пользователях.
type DB struct {
	*mongo.DB // соединение с MongoDB
}

// InitDB возваращает инициализированный объект для работы с хранилищем информации
// о зарегистрированных пользователях системы.
func InitDB(mdb *mongo.DB) (db *DB, err error) {
	db = &DB{mdb}
	return
}

// User описывает информацию о пользователе системы.
type User struct {
	ID      bson.ObjectId `bson:"_id"` // уникальный идентификатор пользователя
	GroupID string        // уникальный идентификатор группы (UUID)
	Name    string        `bson:",omitempty" json:",omitempty"` // отображаемое имя
	Icon    byte          // идентификатор иконки пользователя
}

// Save сохраняет информацию о пользователях в хранилище.
// Если пользователь с таким идентификатором уже существовал, то его данные обновятся.
// Если пользователя с таким идентификатором не существует или он пустой, то в хранилище
// добавится новый пользователь. В последнем случае так же пользователю будет сгенерирован
// новый уникальный идентификатор.
func (db *DB) Save(users ...*User) (err error) {
	coll := db.GetCollection(CollectionName)
	for _, user := range users {
		if !user.ID.Valid() {
			user.ID = bson.NewObjectId()
		}
		if _, err = coll.UpsertId(user.ID, user); err != nil {
			break
		}
	}
	db.FreeCollection(coll)
	return
}

// Get возвращает информацию о пользователе с указанным идентификатором.
func (db *DB) Get(id bson.ObjectId) (user *User, err error) {
	coll := db.GetCollection(CollectionName)
	user = new(User)
	err = coll.FindId(id).One(user)
	db.FreeCollection(coll)
	return
}

type GroupInfo struct {
	GroupID string   // идентификатор группы
	Users   []string // список идентификаторо пользователей с номером иконки
}

// GetGroup возвращает список идентификаторов всех пользователей, входящих в указанную группу.
// Плюс, к идентификатору пользователя автоматичики добавляется номер иконки.
func (db *DB) GetGroup(groupID string) (info *GroupInfo, err error) {
	coll := db.GetCollection(CollectionName)
	users := make([]*User, 0)
	err = coll.Find(bson.M{"groupid": groupID}).Select(bson.M{"icon": 1}).All(&users)
	db.FreeCollection(coll)
	if err != nil {
		return
	}
	if len(users) == 0 {
		return
	}
	info = &GroupInfo{
		GroupID: groupID,
		Users:   make([]string, len(users)),
	}
	for i, user := range users {
		info.Users[i] = fmt.Sprintf("%s-%d", user.ID.Hex(), user.Icon)
	}
	return
}
