package ublox

import (
	"testing"

	"github.com/mdigger/geotrack/mongo"
)

func TestCache(t *testing.T) {
	mongodb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		t.Fatal(err)
	}
	defer mongodb.Close()

	cache, err := InitCache(mongodb, token)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1000; i++ {
		data, err := cache.Get(pointHome, DefaultProfile)
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Println(data)
		data, err = cache.Get(pointWork, DefaultProfile)
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Println(data)
		_ = data
		// jsondata, err := json.Marshal(data)
		// if err != nil {
		// 	t.Fatal(err)
		// }
		// fmt.Println("json:", string(jsondata))
	}
}
