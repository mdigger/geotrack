package main

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/users"
	"github.com/nats-io/nats"
)

func TestSubscription(t *testing.T) {
	log.Print("Connecting to MongoDB...")
	mdb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	log.Println("Connecting to NATS...")
	nc, err := nats.DefaultOptions.Connect()
	if err != nil {
		log.Println("Error connecting to NATS:", err)
		return
	}
	defer nc.Close()

	// запускаем подписку на получение данных и их обработку
	if err := subscribe(mdb, nc); err != nil {
		log.Println("Initializing error:", err)
		return
	}

	nce, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		t.Fatal(err)
	}

	var (
		timemout = 10 * time.Second
		point    = geo.NewPoint(55.715084, 37.57351)
		deviceID = "test0123456789"
		groupID  = users.SampleGroupID
	)
	{
		fmt.Println("LBS Request")
		var data = new(geo.Point)
		err = nce.Request(serviceNameLBS,
			`864078-35827-010003698-fa-2-1e50-772a-95-1e50-773c-a6-1e50-7728-a1-1e50-7725-92-1e50-772d-90-1e50-7741-90-1e50-7726-88`,
			data, timemout)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("LBS Response: %v\n", data)
	}
	{
		fmt.Println("UBLOX Request")
		var data []byte
		err = nce.Request(serviceNameUblox, point, &data, timemout)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("UBLOX Response: %d data length\n", len(data))
	}
	{
		fmt.Println("IMEI Request")
		var data users.GroupInfo
		err = nce.Request(serviceNameIMEI, deviceID, &data, timemout)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("IMEI Response: %v\n", data)
		groupID = data.GroupID
	}
	{
		fmt.Println("TRACK publish")
		var data = &tracks.TrackData{
			GroupID:  groupID,
			DeviceID: deviceID,
			Time:     time.Now(),
			Point:    point,
		}
		err = nce.Publish(serviceNameTracks, data)
		if err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(time.Second * 2) // ожидаем обработки, иначе не успеет
}
