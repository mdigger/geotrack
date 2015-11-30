package users

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/mdigger/geotrack/mongo"
)

func TestUsers(t *testing.T) {
	mdb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	db, err := InitDB(mdb)
	if err != nil {
		t.Fatal(err)
	}
	groupID := db.GetSampleGroupID()
	usersID, err := db.GetGroup(groupID)
	if err != nil {
		t.Fatal(err)
	}
	var response = &struct {
		GroupID string
		Users   []string
	}{
		GroupID: groupID.String(),
		Users:   usersID,
	}
	data, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}
