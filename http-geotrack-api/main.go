package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	logger "github.com/labstack/gommon/log"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/places"
	"github.com/mdigger/geotrack/sensors"
	"github.com/mdigger/geotrack/token"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/users"
	"github.com/nats-io/nats"
)

var (
	e           *echo.Echo
	usersDB     *users.DB             // хранилище пользователей
	placesDB    *places.DB            // хранилище мест
	tracksDB    *tracks.DB            // хранилище треков
	sensorsDB   *sensors.DB           // хранилище сенсоров
	groupID     = users.SampleGroupID // уникальный идентификатор группы
	tokenEngine *token.Engine         // генератор токенов
	nce         *nats.EncodedConn     // соединение с NATS
	llog        *logger.Logger        // вывод информации в лог
)

func main() {
	addr := flag.String("http", ":8080", "Server address & port")
	mongoURL := flag.String("mongodb", "mongodb://localhost/watch", "MongoDB connection URL")
	natsURL := flag.String("nats", nats.DefaultURL, "NATS connection URL")
	docker := flag.Bool("docker", false, "for docker")
	flag.Parse()

	// Если запускается внутри контейнера
	if *docker {
		tmp1 := os.Getenv("NATSADDR")
		tmp2 := os.Getenv("MONGODB")
		natsURL = &tmp1
		mongoURL = &tmp2
	}

	e = echo.New()     // инициализируем HTTP-обработку
	e.Debug()          // режим отладки
	e.SetLogPrefix("") // убираем префикс в логе
	e.SetLogLevel(0)   // устанавливаем уровень вывода всех сообщений (TRACE)
	llog = e.Logger()  // интерфейс вывода в лог

	llog.Info("Connecting to MongoDB %q...", *mongoURL)
	mdb, err := mongo.Connect(*mongoURL)
	if err != nil {
		llog.Error("Error connecting to MongoDB: %v", err)
		return
	}
	defer mdb.Close()
	if usersDB, err = users.InitDB(mdb); err != nil {
		llog.Error("Error initializing UsersDB: %v", err)
		return
	}
	if placesDB, err = places.InitDB(mdb); err != nil {
		llog.Error("Error initializing PlacesDB: %v", err)
		return
	}
	if tracksDB, err = tracks.InitDB(mdb); err != nil {
		llog.Error("Error initializing TracksDB: %v", err)
		return
	}
	if sensorsDB, err = sensors.InitDB(mdb); err != nil {
		llog.Error("Error initializing SensorsDB: %v", err)
		return
	}
	groupID = usersDB.GetSampleGroupID() // временная инициализация пользователей

	log.Println("Connecting to NATS...")
	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Printf("Error connecting to NATS: %v", err)
		return
	}
	defer nc.Close()
	nce, err = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Printf("Error initializing NATS encoder: %v", err)
		return
	}

	// инициализируем работу с токенами
	tokenEngine, err = token.Init("com.xyzrd.geotracker", time.Minute*30, nil)
	if err != nil {
		llog.Error("Error initializing Token Engine: %v", err)
		return
	}
	llog.Debug("CryptoKey: %v", tokenEngine.CryptoKey())

	e.Use(Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	apiV1 := e.Group("/api/v1") // группа URL для обработки API версии 1.0.
	apiV1.Get("/login", login)  // авторизация пользователя

	apiV1Sec := apiV1.Group("") // группа запросов с авторизацией
	apiV1Sec.Use(auth)          // добавляем проверку токена в заголовке

	apiV1Sec.Get("/users", getUsers) // возвращает список пользователей

	apiV1Sec.Get("/places", getPlaces)                // возвращает список интересующих мест
	apiV1Sec.Post("/places", postPlace)               // добавляет определение нового места
	apiV1Sec.Get("/places/:place-id", getPlace)       // возвращает информацию об указаном месте
	apiV1Sec.Put("/places/:place-id", putPlace)       // изменяет определение места
	apiV1Sec.Delete("/places/:place-id", deletePlace) // удаляет определение места

	apiV1Sec.Get("/devices", getDevices)                      // возвращает список устройств
	apiV1Sec.Post("/devices", postDevicePairing)              // привязка устройства к группе
	apiV1Sec.Post("/devices/:device-id", postDevicePairing)   // привязка устройства к группе
	apiV1Sec.Get("/devices/:device-id/tracks", getTracks)     // возвращает список трекингов устройства
	apiV1Sec.Post("/devices/:device-id/tracks", postTracks)   // добавляет данные о треках устройства
	apiV1Sec.Get("/devices/:device-id/sensors", getSensors)   // возвращает список трекингов устройства
	apiV1Sec.Post("/devices/:device-id/sensors", postSensors) // добавляет данные о треках устройства

	apiV1Sec.Post("/push/:push-type", postRegister)            // регистрирует устройство для отправки push-сообщений
	apiV1Sec.Delete("/push/:push-type/:token", deleteRegister) // удаляет токен из хранилища

	llog.Info("Starting HTTP server at %q...", *addr)
	e.Run(*addr)
}
