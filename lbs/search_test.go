package lbs

import (
	"fmt"
	"log"
	"testing"

	"github.com/mdigger/geotrack/mongo"
)

func TestSearch(t *testing.T) {
	mongodb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mongodb.Close()

	lbs, err := InitDB(mongodb)
	if err != nil {
		t.Fatal(err)
	}

	point, err := lbs.SearchLBS(reqStr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(point)
}
