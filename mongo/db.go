package mongo

import "gopkg.in/mgo.v2"

// DB описывает соединение и работу с хранилищем данных.
type DB struct {
	session *mgo.Session // соединение с хранилищем
	dbname  string
}

// Connect устанавливает соединение с хранилищем данных и возвращает его.
func Connect(url string) (db *DB, err error) {
	dialInfo, err := mgo.ParseURL(url)
	if err != nil {
		return
	}
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return
	}
	session.SetMode(mgo.Monotonic, true) // устанавливаем режим запросов
	db = &DB{
		session: session,
		dbname:  dialInfo.Database,
	}
	return
}

// Close закрывает соединение с хранилищем данных.
func (db *DB) Close() {
	db.session.Close()
}

// GetCollection возвращает ссылку на коллекцию с указанным именем. При этом
// автоматически создает копию соединения. По окончании работы с этой коллекцией
// необходимо ее освободить методом FreeCollection().
func (db *DB) GetCollection(name string) *mgo.Collection {
	return db.session.Copy().DB(db.dbname).C(name)
}

// FreeCollection освобождает сессию работы с ранее полученной методом GetCollection()
// коллекцией данных хранилища.
func (db *DB) FreeCollection(coll *mgo.Collection) {
	coll.Database.Session.Close()
}
