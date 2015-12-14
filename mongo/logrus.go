package mongo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	CollectionName = "err.logs"
	ExpireAfter    = time.Duration(time.Hour * 24 * 7)
)

func (db DB) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		// logrus.InfoLevel,
		// logrus.DebugLevel,
	}
}

type logEntry struct {
	Time    time.Time
	Level   string
	Message string        `bson:",omitempty"`
	Data    logrus.Fields `bson:",omitempty"`
	File    string        `bson:",omitempty"`
	Line    int           `bson:",omitempty"`
	// Stack   string        `bson:",omitempty"`
}

func (db *DB) Fire(entry *logrus.Entry) error {
	e := logEntry{
		Time:    entry.Time,
		Level:   entry.Level.String(),
		Message: entry.Message,
		Data:    make(logrus.Fields, len(entry.Data)),
	}
	for name, value := range entry.Data {
		switch t := value.(type) {
		case error:
			e.Data[name] = t.Error()
		// case fmt.Stringer:
		// 	e.Data[name] = t.String()
		default:
			e.Data[name] = value
		}
	}
	if _, file, line, ok := runtime.Caller(4); ok {
		e.File = file
		e.Line = line
	}
	// const size = 64 << 10                            // Размер буфера под стек
	// buf := make([]byte, size)                        // Инициализируем буфер под загрузку стека
	// e.Stack = string(buf[:runtime.Stack(buf, true)]) // Получаем стек ошибки

	coll := db.GetCollection(CollectionName)
	mgoErr := coll.Insert(e)
	db.FreeCollection(coll)
	if mgoErr != nil {
		return fmt.Errorf("Failed to send log entry to mongodb: %s", mgoErr)
	}
	return nil
}
