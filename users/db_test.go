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
	groupInfo, err := db.GetGroup(groupID)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(groupInfo, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
}
