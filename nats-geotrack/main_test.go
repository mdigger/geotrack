package main

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mdigger/geolocate"
	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/lbs/parser"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/sensors"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/users"
	"github.com/nats-io/nats"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// log.SetFormatter(&log.JSONFormatter{})
	// Output to stderr instead of stdout, could also be a file.
	// log.SetOutput(os.Stderr)
	// Only log the warning severity or above.
	// log.SetLevel(log.WarnLevel)
}

func TestSubscription(t *testing.T) {
	log.SetLevel(log.DebugLevel) // отладка

	log.WithField("url", "mongodb://localhost/watch").Info("Connecting to MongoDB...")
	mdb, err := mongo.Connect("mongodb://localhost/watch")
	if err != nil {
		log.WithError(err).Error("Error connecting to MongoDB")
		return
	}
	defer mdb.Close()
	log.AddHook(mdb) // добавляем запись логов в MongoDB

	log.WithField("options", nats.DefaultOptions.Url).Info("Connecting to NATS...")
	nc, err := nats.DefaultOptions.Connect()
	if err != nil {
		log.WithError(err).Error("Error connecting to NATS")
		return
	}
	defer nc.Close()

	// запускаем подписку на получение данных и их обработку
	if err := subscribe(mdb, nc); err != nil {
		log.WithError(err).Error("Error initializing NATS subscription")
		return
	}

	nce, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.WithError(err).Fatal("Error initializing NATS encoder")
	}

	var (
		timemout = 10 * time.Second
		point    = geo.Point{55.715084, 37.57351}
		deviceID = "test0123456789"
		groupID  = users.SampleGroupID
	)
	{
		req, err := parser.ParseLBS("gsm", `3D867293-13007-022970357-fa-2-1e3f-57f5-8b-1e3f-9b10-8a-1e3f-5100-79-1e3f-b6a6-78-1e3f-6aaa-78-1e3f-57f6-77-1e3f-5103-72`)
		if err != nil {
			log.WithError(err).Fatal("LBS Parse")
		}
		var data geolocate.Response
		err = nce.Request(serviceNameLBS, *req, &data, timemout)
		if err != nil {
			log.WithError(err).Fatal("LBS Request")
		}
	}
	{
		var data []byte
		err = nce.Request(serviceNameUblox, point, &data, timemout)
		if err != nil {
			log.WithError(err).Fatal("UBLOX Request")
		}
	}
	{
		var data users.GroupInfo
		err = nce.Request(serviceNameIMEI, deviceID, &data, timemout)
		if err != nil {
			log.WithError(err).Fatal("IMEI Request")
		}
		groupID = data.GroupID
	}
	{
		var data = tracks.TrackData{
			GroupID:  groupID,
			DeviceID: deviceID,
			Time:     time.Now(),
			Location: point,
		}
		err = nce.Publish(serviceNameTracks, []tracks.TrackData{data})
		if err != nil {
			log.WithError(err).Fatal("TRACKS publishing")
		}
	}
	{
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
			log.WithError(err).Fatal("SENSORS publishing")
		}
	}
	{
		var key string
		err = nce.Request(serviceNamePairing, deviceID, &key, timemout)
		if err != nil {
			log.WithError(err).Fatal("PAIRING Request")
		}
		if key == "" {
			log.Fatal("empty key")
		}

		var newDeviceID string
		err = nce.Request(serviceNamePairingKey, key, &newDeviceID, timemout)
		if err != nil {
			log.WithError(err).Fatal("PAIRING KEY Request")
		}
		if newDeviceID == "" {
			log.Fatal("bad device pairing deviceid")
		}
		if newDeviceID != deviceID {
			log.Fatal("bad device pairing")
		}
	}

	time.Sleep(time.Second * 5) // ожидаем обработки, иначе не успеет
}
