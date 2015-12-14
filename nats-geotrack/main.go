package main

import (
	"flag"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
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
	serviceNameLBS        = "service.lbs"
	serviceNameUblox      = "service.ublox"
	serviceNameIMEI       = "device.imei"
	serviceNameTracks     = "device.track"
	serviceNameSensors    = "device.sensor"
	serviceNamePairing    = "device.pair"
	serviceNamePairingKey = "device.pair.key"
)

var (
	ubloxToken  = "I6KKO4RU_U2DclBM9GVyrA"
	googleToken = "AIzaSyBDw1oDEngRh098SlWFKWDJ5k7BFrfX_WI"
)

func main() {
	log.SetLevel(log.DebugLevel) // отладка

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

	log.WithField("url", *mongoURL).Info("Connecting to MongoDB...")
	mdb, err := mongo.Connect(*mongoURL)
	if err != nil {
		log.WithError(err).Error("Error connecting to MongoDB")
		return
	}
	defer mdb.Close()
	log.AddHook(mdb) // добавляем запись логов в MongoDB

	log.WithField("url", *natsURL).Info("Connecting to NATS...")
	nc, err := nats.Connect(*natsURL)
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
	// блокируем дальнейший код до получения одного из сигналов
	monitorSignals(os.Interrupt, os.Kill)
	log.Debug("THE END")
}

func subscribe(mdb *mongo.DB, nc *nats.Conn) error {
	nce, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	// nce.Subscribe("*", func(subj, reply string, data []byte) {
	// 	log.Printf("DEBUG: %q [%q]\n%s", subj, reply, string(data))
	// })

	lbs, err := lbs.InitDB(mdb)
	if err != nil {
		return err
	}
	if lbs.Records() == 0 {
		log.Warn("LBS DB is empty!")
	}
	lbsGoogle, err := geolocate.New(geolocate.Google, googleToken)
	if err != nil {
		return err
	}
	nce.Subscribe(serviceNameLBS, func(_, reply string, req geolocate.Request) {
		resp, err := lbsGoogle.Get(req)
		logger := log.WithFields(log.Fields{"request": req, "response": resp})
		if err != nil {
			logger.WithError(err).Error("LBS Google error")
		}
		if err := nce.Publish(reply, resp); err != nil {
			logger.WithError(err).Error("LBS Google response error")
		} else {
			logger.Debug("LBS")
		}
		_, err = lbs.Get(req) // оставил для сохранения в базу запроса.
		if err != nil {
			logger.WithError(err).Error("LBS Internal response error")
		}
	})

	ubloxCache, err := ublox.InitCache(mdb, ubloxToken)
	if err != nil {
		return err
	}
	profile := ublox.DefaultProfile
	nce.Subscribe(serviceNameUblox, func(_, reply string, point geo.Point) {
		data, err := ubloxCache.Get(point, profile)
		logger := log.WithFields(log.Fields{"request": point, "response length": len(data)})
		if err != nil {
			logger.WithError(err).Error("UBLOX error")
		}
		if err := nce.Publish(reply, data); err != nil {
			logger.WithError(err).Error("UBLOX response error")
		} else {
			logger.Debug("UBLOX")
		}
	})

	usersDB, err := users.InitDB(mdb)
	if err != nil {
		return err
	}
	// уникальный идентификатор группы пока для примера задан явно
	// groupID := users.SampleGroupID
	groupID := usersDB.GetSampleGroupID()
	nce.Subscribe(serviceNameIMEI, func(_, reply, data string) {
		group, err := usersDB.GetGroup(groupID)
		logger := log.WithFields(log.Fields{"request": data, "response": group})
		if err != nil {
			logger.WithError(err).Error("IMEI error")
		}
		if err := nce.Publish(reply, group); err != nil {
			logger.WithError(err).Error("IMEI response error")
		} else {
			logger.Debug("IMEI")
		}
	})

	tracksDB, err := tracks.InitDB(mdb)
	if err != nil {
		return err
	}
	nce.Subscribe(serviceNameTracks, func(tracks []tracks.TrackData) {
		logger := log.WithField("request", tracks)
		if err := tracksDB.Add(tracks...); err != nil {
			logger.WithError(err).Error("TRACKS error")
		} else {
			logger.Debug("TRACKS")
		}
	})

	sensorsDB, err := sensors.InitDB(mdb)
	if err != nil {
		return err
	}
	nce.Subscribe(serviceNameSensors, func(sensors []sensors.SensorData) {
		logger := log.WithField("request", sensors)
		if err := sensorsDB.Add(sensors...); err != nil {
			logger.WithError(err).Error("SENSORS error")
		} else {
			logger.Debug("SENSORS")
		}
	})

	var pairs pairing.Pairs
	nce.Subscribe(serviceNamePairing, func(_, reply, deviceID string) {
		key := pairs.Generate(deviceID)
		logger := log.WithFields(log.Fields{"request": deviceID, "response": key})
		if err := nce.Publish(reply, key); err != nil {
			logger.WithError(err).Error("PAIR error")
		} else {
			logger.Debug("PAIR")
		}
	})
	nce.Subscribe(serviceNamePairingKey, func(_, reply, key string) {
		newDeviceID := pairs.GetDeviceID(key)
		logger := log.WithFields(log.Fields{"request": key, "response": newDeviceID})
		if err := nce.Publish(reply, newDeviceID); err != nil {
			logger.WithError(err).Error("PAIR KEY error")
		} else {
			logger.Debug("PAIR KEY")
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
