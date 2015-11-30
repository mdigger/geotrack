package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/mdigger/geotrack/geo"
	"github.com/mdigger/geotrack/lbs"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/ublox"
	"github.com/mdigger/geotrack/users"
	"github.com/nats-io/nats"
)

const (
	serviceNameLBS    = "lbs"
	serviceNameUblox  = "ublox"
	serviceNameTracks = "tracks"
	serviceNameIMEI   = "imei"
)

func main() {
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
	// блокируем дальнейший код до получения одного из сигналов
	monitorSignals(os.Interrupt, os.Kill)
	log.Println("THE END")
}

func subscribe(mdb *mongo.DB, nc *nats.Conn) error {
	nce, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	log.Println("Initializing LBS subscription...")
	lbs, err := lbs.InitDB(mdb)
	if err != nil {
		return err
	}
	if lbs.RecordsCount() == 0 {
		log.Println("Warning! LBS DB is empty!")
	}
	nce.Subscribe(serviceNameLBS, func(_, reply, data string) {
		log.Println("LBS:", data)
		point, err := lbs.SearchLBS(data)
		if err != nil {
			log.Println("LBS error:", err)
		}
		if err := nce.Publish(reply, point); err != nil {
			log.Println("LBS reply error:", err)
		}
	})

	log.Println("Initializing UBLOX subscription...")
	ubloxCache, err := ublox.InitCache(mdb, "I6KKO4RU_U2DclBM9GVyrA")
	if err != nil {
		return err
	}
	profile := ublox.DefaultProfile
	nce.Subscribe(serviceNameUblox, func(_, reply string, point *geo.Point) {
		log.Println("UBLOX:", point)
		data, err := ubloxCache.Get(point, profile)
		if err != nil {
			log.Println("UBLOX error:", err)
		}
		if err := nce.Publish(reply, data); err != nil {
			log.Println("UBLOX reply error:", err)
		}
	})

	log.Println("Initializing Tracks subscription...")
	tracksDB, err := tracks.InitDB(mdb)
	if err != nil {
		return err
	}
	nce.Subscribe(serviceNameTracks, func(data *tracks.TrackData) {
		log.Println("Track:", data)
		if err := tracksDB.Add(data); err != nil {
			log.Println("Error TrackDB Add:", err)
		}
	})

	log.Println("Initializing IMEI Identification subscription...")
	usersDB, err := users.InitDB(mdb)
	if err != nil {
		return err
	}
	// уникальный идентификатор группы пока для примера захардкоден
	groupID := usersDB.GetSampleGroupID()
	nce.Subscribe(serviceNameIMEI, func(_, reply, data string) {
		log.Println("IMEI:", data)
		group, err := usersDB.GetGroup(groupID)
		if err != nil {
			log.Println("Error getting group of users:", err)
		}
		if err := nce.Publish(reply, group); err != nil {
			log.Println("IMEI reply error:", err)
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
