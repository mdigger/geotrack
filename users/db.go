package users

import (
	"fmt"

	"github.com/mdigger/geotrack/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var CollectionName = "users" // содержит название коллекции с данными треков

// DB описывает интерфейс для работы с хранилищем данных с информацией о зарегистрированных
// в системе пользователях.
type DB struct {
	*mongo.DB // соединение с MongoDB
}

// InitDB возваращает инициализированный объект для работы с хранилищем информации
// о зарегистрированных пользователях системы.
func InitDB(mdb *mongo.DB) (db *DB, err error) {
	db = &DB{mdb}
	coll := mdb.GetCollection(CollectionName)
	err = coll.EnsureIndex(mgo.Index{
		Key:    []string{"login"},
		Unique: true,
	})
	if err != nil {
		return
	}
	err = coll.EnsureIndexKey("groupid")
	if err != nil {
		return
	}
	mdb.FreeCollection(coll)
	return
}

// User описывает информацию о пользователе системы.
type User struct {
	ID       bson.ObjectId `bson:"_id"` // уникальный идентификатор пользователя
	Login    string        // логин пользователя
	GroupID  string        `json:",omitempty"`                   // уникальный идентификатор группы (UUID)
	Name     string        `bson:",omitempty" json:",omitempty"` // отображаемое имя
	Icon     byte          // идентификатор иконки пользователя
	Password []byte        `json:"-"` // хеш пароля пользователя
}

// Get возвращает информацию о пользователе с указанным идентификатором.
func (db *DB) Get(login string) (user *User, err error) {
	coll := db.GetCollection(CollectionName)
	user = new(User)
	err = coll.Find(bson.M{"login": login}).One(user)
	db.FreeCollection(coll)
	return
}

// Check возвращает true, если пользователь с таким идентификатором действительно существует и
// находится в данной группе.
func (db *DB) Check(groupID string, userID bson.ObjectId) (exists bool, err error) {
	coll := db.GetCollection(CollectionName)
	count, err := coll.Find(bson.M{"_id": userID, "groupid": groupID}).Count()
	db.FreeCollection(coll)
	exists = (count == 1)
	return
}

// Save сохраняет информацию о пользователях в хранилище.
// Если пользователь с таким идентификатором уже существовал, то его данные обновятся.
// Если пользователя с таким идентификатором не существует или он пустой, то в хранилище
// добавится новый пользователь. В последнем случае так же пользователю будет сгенерирован
// новый уникальный идентификатор.
func (db *DB) Save(user *User) (err error) {
	coll := db.GetCollection(CollectionName)
	defer db.FreeCollection(coll)
	if !user.ID.Valid() {
		user.ID = bson.NewObjectId()
	}
	_, err = coll.UpsertId(user.ID, user)
	if err != nil {
		return
	}
	return
}

type GroupInfo struct {
	GroupID string   // идентификатор группы
	Users   []string // список идентификаторов пользователей с номером иконки
}

// GetGroup возвращает список идентификаторов всех пользователей, входящих в указанную группу.
// Плюс, к идентификатору пользователя автоматически добавляется номер иконки.
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

// GetUsers возвращает список всех пользователей, входящих в указанную группу.
func (db *DB) GetUsers(groupID string) (users []*User, err error) {
	coll := db.GetCollection(CollectionName)
	users = make([]*User, 0)
	selector := bson.M{"groupid": 0, "password": 0}
	err = coll.Find(bson.M{"groupid": groupID}).Select(selector).All(&users)
	db.FreeCollection(coll)
	return
}
