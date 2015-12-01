package users

import "log"

var SampleGroupID = "540da544-981c-11e5-a22e-28cfe91a86a7"

// GetSampleGroupID возвращает тестовый предопределенный идентификатор группы.
func (db *DB) GetSampleGroupID() string {
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
