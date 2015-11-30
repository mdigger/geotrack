package users

import (
	"log"

	"github.com/mdigger/uuid"
)

var SampleGroupID = uuid.UUID{32, 108, 89, 30, 161, 81, 69, 64, 189, 203, 0, 195, 95, 149, 121, 43}

// GetSampleGroupID возвращает тестовый предопределенный идентификатор группы.
func (db *DB) GetSampleGroupID() uuid.UUID {
	group, err := db.GetGroup(SampleGroupID)
	if err != nil {
		panic(err)
	}
	// наполняем тестовыми данными
	if group == nil {
		log.Println("USERS: Generating sample users...")
		for i := 0; i < 5; i++ {
			err := db.Save(&User{
				GroupID: SampleGroupID,
				Icon:    byte(i),
			})
			if err != nil {
				panic(err)
			}
		}
	}
	return SampleGroupID
}
