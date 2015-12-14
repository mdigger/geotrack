package users

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var SampleGroupID = "540da544-981c-11e5-a22e-28cfe91a86a7"

// GetSampleGroupID возвращает тестовый предопределенный идентификатор группы.
func (db *DB) GetSampleGroupID() string {
	group, err := db.GetGroup(SampleGroupID)
	if err != nil {
		panic(err)
	}
	// наполняем тестовыми данными
	if group == nil {
		for i := 0; i < 5; i++ {
			password, err := bcrypt.GenerateFromPassword(
				[]byte(fmt.Sprintf("password%d", i+1)), bcrypt.DefaultCost)
			if err != nil {
				panic(err)
			}
			err = db.Save(&User{
				Login:    fmt.Sprintf("login%d", i+1),
				GroupID:  SampleGroupID,
				Name:     fmt.Sprintf("User #%d", i+1),
				Icon:     byte(i),
				Password: password,
			})
			if err != nil {
				panic(err)
			}
		}
	}
	return SampleGroupID
}
