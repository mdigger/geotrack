package lbs

import (
	"fmt"
	"log"
	"testing"

	"github.com/mdigger/geolocate"
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

	request := geolocate.Request{
		CellTowers: []geolocate.CellTower{
			{250, 2, 7743, 22517, -78, 0, 0},
			{250, 2, 7743, 39696, -81, 0, 0},
			{250, 2, 7743, 22518, -91, 0, 0},
			{250, 2, 7743, 27306, -101, 0, 0},
			{250, 2, 7743, 29909, -103, 0, 0},
			{250, 2, 7743, 22516, -104, 0, 0},
			{250, 2, 7743, 20736, -105, 0, 0},
		},
	}

	cells, err := lbs.GetCells(request)
	if err != nil {
		t.Fatal(err)
	}
	for _, cell := range cells {
		fmt.Println(cell)
	}
	resp, err := lbs.Get(request)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(resp)
}
