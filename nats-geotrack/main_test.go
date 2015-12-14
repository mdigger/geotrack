package main

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/mdigger/geolocate"
	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/lbs/parser"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/sensors"
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
		point    = geo.Point{55.715084, 37.57351}
		deviceID = "test0123456789"
		groupID  = users.SampleGroupID
	)
	{
		fmt.Println("LBS Request")
		req, err := parser.ParseLBS("gsm", `3D867293-13007-022970357-fa-2-1e3f-57f5-8b-1e3f-9b10-8a-1e3f-5100-79-1e3f-b6a6-78-1e3f-6aaa-78-1e3f-57f6-77-1e3f-5103-72`)
		if err != nil {
			t.Fatal(err)
		}
		var data geolocate.Response
		err = nce.Request(serviceNameLBS, *req, &data, timemout)
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
		var data = tracks.TrackData{
			GroupID:  groupID,
			DeviceID: deviceID,
			Time:     time.Now(),
			Location: point,
		}
		err = nce.Publish(serviceNameTracks, []tracks.TrackData{data})
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		fmt.Println("SENSOR publish")
		var data = sensors.SensorData{
			GroupID:  groupID,
			DeviceID: deviceID,
			Time:     time.Now(),
			Data: map[string]interface{}{
				"sensor1": uint(1),
				"sensor2": "sensor 2",
				"sensor3": []uint{1, 2, 3},
			},
		}
		err = nce.Publish(serviceNameSensors, []sensors.SensorData{data})
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		fmt.Println("PAIRING Request")
		var data string
		err = nce.Request(serviceNamePairing, deviceID, &data, timemout)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("PAIRING Response: %v\n", data)
	}

	time.Sleep(time.Second * 5) // ожидаем обработки, иначе не успеет
}
