package parser

import (
	"fmt"
	"testing"

	"github.com/mdigger/geolocate"
	"github.com/mdigger/geotrack/lbs"
	"github.com/mdigger/geotrack/mongo"
)

const (
	// googleAPIKey = "<API Key>"
	// yandexAPIKey = "<API Key>"
	googleAPIKey = "AIzaSyBDw1oDEngRh098SlWFKWDJ5k7BFrfX_WI"
	yandexAPIKey = "AFVzaFYBAAAAtuSxOwMARr9H5Bte6G1FlAeX4vHlmPYgQeEAAAAAAAAAAABiS9nunmKQsVptuGIh0rAM1K8PoQ=="
)

func TestLocators(t *testing.T) {
	google, err := geolocate.New(geolocate.Google, googleAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	mozilla, err := geolocate.New(geolocate.Mozilla, "test")
	if err != nil {
		t.Fatal(err)
	}
	yandex, err := geolocate.New(geolocate.Yandex, yandexAPIKey)
	if err != nil {
		t.Fatal(err)
	}
	mongodb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		t.Fatal(err)
	}
	defer mongodb.Close()
	lbs, err := lbs.InitDB(mongodb)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range testData {
		req, err := ParseLBS("gsm", s)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := google.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Google", err)
			t.Error(err)
		} else {
			fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Google", resp.Location.Lng, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = yandex.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Yandex", err)
			t.Error(err)
		} else {
			fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Yandex", resp.Location.Lng, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = mozilla.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Mozilla", err)
			t.Error(err)
		} else {
			fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Mozilla", resp.Location.Lng, resp.Location.Lat, resp.Accuracy)
		}
		resp, err = lbs.Get(*req)
		if err != nil {
			fmt.Printf("%8s error: %v\n", "Internal", err)
			t.Error(err)
		}
		fmt.Printf("%8s [%.7f,%.7f] %.2f\n", "Internal", resp.Location.Lng, resp.Location.Lat, resp.Accuracy)
		fmt.Println("-----------------------------------------")
	}
}
