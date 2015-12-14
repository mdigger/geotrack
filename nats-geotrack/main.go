package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/mdigger/geolocate"
	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/lbs"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/pairing"
	"github.com/mdigger/geotrack/sensors"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/ublox"
	"github.com/mdigger/geotrack/users"
	"github.com/nats-io/nats"
)

const (
	serviceNameLBS        = "lbs"
	serviceNameUblox      = "ublox"
	serviceNameIMEI       = "imei"
	serviceNameTracks     = "track"
	serviceNameSensors    = "sensor"
	serviceNamePairing    = "pairing"
	serviceNamePairingKey = "pairing.key"
)

var (
	ubloxToken  = "I6KKO4RU_U2DclBM9GVyrA"
	googleToken = "AIzaSyBDw1oDEngRh098SlWFKWDJ5k7BFrfX_WI"
)

func main() {
	mongoURL := flag.String("mongodb", "mongodb://localhost/watch", "MongoDB connection URL")
	natsURL := flag.String("nats", nats.DefaultURL, "NATS connection URL")
	docker := flag.Bool("docker", false, "for docker")
	flag.StringVar(&ubloxToken, "ublox", ubloxToken, "U-Blox token")
	flag.Parse()

	// Если запускается внутри контейнера
	if *docker {
		tmp1 := os.Getenv("NATSADDR")
		tmp2 := os.Getenv("MONGODB")
		natsURL = &tmp1
		mongoURL = &tmp2
	}

	log.Print("Connecting to MongoDB...")
	mdb, err := mongo.Connect(*mongoURL)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		return
	}
	defer mdb.Close()

	log.Println("Connecting to NATS...")
	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Printf("Error connecting to NATS: %v", err)
		return
	}
	defer nc.Close()

	// запускаем подписку на получение данных и их обработку
	if err := subscribe(mdb, nc); err != nil {
		log.Printf("Initializing error: %v", err)
		return
	}
	// блокируем дальнейший код до получения одного из сигналов
	monitorSignals(os.Interrupt, os.Kill)
	log.Println("THE END")
}

func subscribe(mdb *mongo.DB, nc *nats.Conn) error {
	nce, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	// nce.Subscribe("*", func(subj, reply string, data []byte) {
	// 	log.Printf("DEBUG: %q [%q]\n%s", subj, reply, string(data))
	// })

	log.Println("Initializing LBS subscription...")
	lbs, err := lbs.InitDB(mdb)
	if err != nil {
		return err
	}
	if lbs.Records() == 0 {
		log.Println("Warning! LBS DB is empty!")
	}
	lbsGoogle, err := geolocate.New(geolocate.Google, googleToken)
	if err != nil {
		return err
	}
	nce.Subscribe(serviceNameLBS, func(_, reply string, req geolocate.Request) {
		log.Printf("LBS:  %+v", req)
		resp, err := lbsGoogle.Get(req)
		if err != nil {
			log.Printf("LBS Google error: %v", err)
		}
		log.Printf("LBS Google Response: %+v", resp)
		if err := nce.Publish(reply, resp); err != nil {
			log.Printf("LBS reply error:  %v [%+v]", err, resp)
		}
		_, err = lbs.Get(req) // оставил для сохранения в базу запроса.
		if err != nil {
			log.Printf("LBS internal error: %v", err)
		}
	})

	log.Println("Initializing UBLOX subscription...")
	ubloxCache, err := ublox.InitCache(mdb, ubloxToken)
	if err != nil {
		return err
	}
	profile := ublox.DefaultProfile
	nce.Subscribe(serviceNameUblox, func(_, reply string, point geo.Point) {
		log.Printf("UBLOX: %v", point)
		data, err := ubloxCache.Get(point, profile)
		if err != nil {
			log.Printf("UBLOX error: %v", err)
		}
		if err := nce.Publish(reply, data); err != nil {
			log.Printf("UBLOX reply error:  %v [%+v]", err, data)
		}
	})

	log.Println("Initializing IMEI Identification subscription...")
	usersDB, err := users.InitDB(mdb)
	if err != nil {
		return err
	}
	// уникальный идентификатор группы пока для примера задан явно
	// groupID := users.SampleGroupID
	groupID := usersDB.GetSampleGroupID()
	nce.Subscribe(serviceNameIMEI, func(_, reply, data string) {
		log.Printf("IMEI: %v", data)
		group, err := usersDB.GetGroup(groupID)
		if err != nil {
			log.Printf("Error getting group of users: %v", err)
		}
		if err := nce.Publish(reply, group); err != nil {
			log.Printf("IMEI reply error: %v [%+v]", err, group)
		}
	})

	log.Println("Initializing Tracks subscription...")
	tracksDB, err := tracks.InitDB(mdb)
	if err != nil {
		return err
	}
	nce.Subscribe(serviceNameTracks, func(tracks []tracks.TrackData) {
		log.Printf("TRACK: %v", tracks)
		if err := tracksDB.Add(tracks...); err != nil {
			log.Printf("Error TrackDB Add:  %v [%+v]", err, tracks)
		}
	})

	log.Println("Initializing Sensors subscription...")
	sensorsDB, err := sensors.InitDB(mdb)
	if err != nil {
		return err
	}
	nce.Subscribe(serviceNameSensors, func(sensors []sensors.SensorData) {
		log.Printf("SENSORS: %v", sensors)
		if err := sensorsDB.Add(sensors...); err != nil {
			log.Printf("Error SensorDB Add: %v [%+v]", err, sensors)
		}
	})

	log.Println("Initializing Pairing subscription...")
	var pairs pairing.Pairs
	nce.Subscribe(serviceNamePairing, func(_, reply, deviceID string) {
		key := pairs.Generate(deviceID)
		log.Printf("PAIRING: %v = %q", deviceID, key)
		if err := nce.Publish(reply, key); err != nil {
			log.Printf("PAIRING reply error: %v [%+v]", err, key)
		}
	})
	nce.Subscribe(serviceNamePairingKey, func(_, reply, key string) {
		newDeviceID := pairs.GetDeviceID(key)
		log.Printf("PAIRING KEY: %v = %q", key, newDeviceID)
		if err := nce.Publish(reply, newDeviceID); err != nil {
			log.Printf("PAIRING KEY reply error: %v [%+v]", err, newDeviceID)
		}
	})

	return nil
}

// monitorSignals запускает мониторинг сигналов и возвращает значение, когда получает сигнал.
// В качестве параметров передается список сигналов, которые нужно отслеживать.
func monitorSignals(signals ...os.Signal) os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)
	return <-signalChan
}
