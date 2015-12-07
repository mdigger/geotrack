package main

import (
	"encoding/hex"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	logger "github.com/labstack/gommon/log"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/places"
	"github.com/mdigger/geotrack/token"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/users"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
)

var (
	usersDB     *users.DB             // хранилище пользователей
	placesDB    *places.DB            // хранилище мест
	tracksDB    *tracks.DB            // хранилище треков
	groupID     = users.SampleGroupID // уникальный идентификатор группы
	tokenEngine *token.Engine         // генератор токенов
	log         *logger.Logger        // вывод информации в лог
)

const (
	tracksLimit = 200 // лимит при отдаче списка треков
)

func main() {
	addr := flag.String("http", ":8080", "Server address & port")
	mongoURL := flag.String("mongodb", "mongodb://localhost/watch", "MongoDB connection URL")
	docker := flag.Bool("docker", false, "for docker")
	flag.Parse()

	// Если запускается внутри контейнера
	if *docker {
		tmp := os.Getenv("MONGODB")
		mongoURL = &tmp
	}

	e := echo.New()    // инициализируем HTTP-обработку
	e.Debug()          // режим отладки
	e.SetLogPrefix("") // убираем префикс в логе
	e.SetLogLevel(0)   // устанавливаем уровень вывода всех сообщений (TRACE)
	log = e.Logger()   // интерфейс вывода в лог

	log.Info("Connecting to MongoDB %q...", *mongoURL)
	mdb, err := mongo.Connect(*mongoURL)
	if err != nil {
		log.Error("Error connecting to MongoDB: %v", err)
		return
	}
	defer mdb.Close()
	if usersDB, err = users.InitDB(mdb); err != nil {
		log.Error("Error initializing UsersDB: %v", err)
		return
	}
	if placesDB, err = places.InitDB(mdb); err != nil {
		log.Error("Error initializing PlacesDB: %v", err)
		return
	}
	if tracksDB, err = tracks.InitDB(mdb); err != nil {
		log.Error("Error initializing TracksDB: %v", err)
		return
	}
	groupID = usersDB.GetSampleGroupID() // временная инициализация пользователей

	// инициализируем работу с токенами
	tokenEngine, err = token.Init("com.xyzrd.geotracker", time.Minute*30, nil)
	if err != nil {
		log.Error("Error initializing Token Engine: %v", err)
		return
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	apiV1 := e.Group("/api/v1")                                    // группа URL для обработки API версии 1.0.
	apiV1.Get("/login", login)                                     // авторизация пользователя
	apiV1Sec := apiV1.Group("")                                    // группа запросов с авторизацией
	apiV1Sec.Use(auth)                                             // добавляем проверку токена в заголовке
	apiV1Sec.Get("/users", getUserslList)                          // возвращает список пользователей
	apiV1Sec.Get("/places", getPlaceslList)                        // возвращает список интересующих мест
	apiV1Sec.Get("/devices", getDeviceslList)                      // возвращает список устройств
	apiV1Sec.Get("/devices/:device-id", getDeviceCurrent)          // возвращает последнюю точку трекинга устройства
	apiV1Sec.Get("/devices/:device-id/history", getDeviceHistory)  // возвращает список треков устройства
	apiV1Sec.Post("/register/:push-type", postRegister)            // регистрирует устройство для отправки push-сообщений
	apiV1Sec.Delete("/register/:push-type/:token", deleteRegister) // удаляет токен из хранилища

	log.Info("Starting HTTP server at %q...", *addr)
	e.Run(*addr)
}

// login читает заголовок запроса с HTTP Basic авторизацией, проверяет пользователя
// по базе данных и отдает в ответ авторизационный ключ в формате JWT.
func login(c *echo.Context) error {
	// получаем пароль из заголовка HTTP Basic авторизации
	username, password, ok := c.Request().BasicAuth()
	if !ok {
		c.Response().Header().Set(echo.WWWAuthenticate, "Basic realm=Restricted")
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	// получаем из хранилища информацию о пользователе
	user, err := usersDB.Get(username)
	switch err {
	case mgo.ErrNotFound: // пользователь не найден
		return echo.NewHTTPError(http.StatusForbidden)
	case nil: // пользователь найден — продолжаем
		break
	default: // другая ошибка
		return err
	}
	// сравниваем сохраненный пароль с тем, что указали в заголовке
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return echo.NewHTTPError(http.StatusForbidden)
	}
	// генерируем JWT-токен
	tokenString, err := tokenEngine.Token(map[string]interface{}{
		"id":    user.ID,
		"group": user.GroupID,
	})
	if err != nil {
		return err
	}
	// отдаем в ответ сервера
	response := c.Response()
	response.Header().Set(echo.ContentType, "application/jwt")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(tokenString))
	return nil
}

// auth является вспомогательной функцией, проверяющей и разбирающей токен с авторизационной
// информацией в HTTP-заголовке. Разобранная информация сохраняется в контексте запроса.
func auth(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req := c.Request()                                  // получаем доступ к HTTP-запросу
		if req.Header.Get(echo.Upgrade) == echo.WebSocket { // пропускаем запросы WebSocket
			return nil
		}
		data, err := tokenEngine.ParseRequest(req) // разбираем токен из запроса
		if err != nil {
			log.Warn("Bad token: %v", err)
			return echo.NewHTTPError(http.StatusForbidden)
		}
		groupID := data["group"].(string)
		userID := data["id"].(string)
		// проверяем, что пользователь есть и входит в эту группу
		exists, err := usersDB.Check(groupID, userID)
		if err != nil {
			return err
		}
		if !exists {
			return echo.NewHTTPError(http.StatusForbidden)
		}
		c.Set("GroupID", groupID) // сохраняем данные в контексте запроса
		c.Set("ID", userID)
		return h(c) // выполняем основной обработчик
	}
}

// getUserslList отдает список зарегистрированных пользователей, которые относятся к той же
// группе, что и текущий пользователь.
func getUserslList(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)    // получаем идентификатор группы
	users, err := usersDB.GetUsers(groupID) // запрашиваем список пользователей
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, users)
}

// getPlaceslList список мест, зарегистрированны для группы пользователей
func getPlaceslList(c *echo.Context) error {
	groupID := c.Get("GroupID").(string) // получаем идентификатор группы
	places, err := placesDB.Get(groupID) // запрашиваем список мест
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, places)
}

// getDeviceslList отдает список зарегистрированных устройств, которые относятся к той же
// группе, что и текущий пользователь.
func getDeviceslList(c *echo.Context) error {
	// TODO: возвращать все устройства, а не только те, треки по которым сохранились
	groupID := c.Get("GroupID").(string)             // получаем идентификатор группы
	deviceIDs, err := tracksDB.GetDevicesID(groupID) // запрашиваем список устройств
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, deviceIDs)
}

// getDeviceCurrent отдает последние данные с координатами браслета.
func getDeviceCurrent(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)              // получаем идентификатор группы
	deviceID := c.Param("device-id")                  // получаем идентификатор устройства
	track, err := tracksDB.GetLast(groupID, deviceID) // запрашиваем список устройств
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, track)
}

// getDeviceHistory отдает всю историю с координатами трекинга браслета, разбивая ее на порции.
func getDeviceHistory(c *echo.Context) error {
	groupID := c.Get("GroupID").(string) // получаем идентификатор группы
	deviceID := c.Param("device-id")     // получаем идентификатор устройства
	lastID := c.Query("last")            // получаем идентификатор последнего полученного трека
	// запрашиваем список устройств постранично
	tracks, err := tracksDB.Get(groupID, deviceID, tracksLimit, lastID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, tracks)
}

// postRegister регистрирует устройство, чтобы на него можно было отсылать push-уведомления.
func postRegister(c *echo.Context) error {
	pushType := c.Param("push-type") // получаем идентификатор типа уведомлений
	switch pushType {
	case "apns": // Apple Push Notification
	case "gcm": // Google Cloud Messages
	default: // не поддерживаемы тип
		return echo.NewHTTPError(http.StatusNotFound)
	}
	groupID := c.Get("GroupID").(string) // идентификатор группы
	userID := c.Get("ID").(string)       // идентификатор пользователя
	token := c.Form("token")             // токен устройства
	if len(token) != 32 {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token size")
	}
	btoken, err := hex.DecodeString(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token format")
	}
	_ = btoken
	return echo.NewHTTPError(http.StatusNotImplemented)
}

// deleteRegister регистрирует устройство, чтобы на него можно было отсылать push-уведомления.
func deleteRegister(c *echo.Context) error {
	pushType := c.Param("push-type") // получаем идентификатор типа уведомлений
	switch pushType {
	case "apns": // Apple Push Notification
	case "gcm": // Google Cloud Messages
	default: // не поддерживаемы тип
		return echo.NewHTTPError(http.StatusNotFound)
	}
	groupID := c.Get("GroupID").(string) // идентификатор группы
	userID := c.Get("ID").(string)       // идентификатор пользователя
	token := c.Param("token")            // токен устройства
	if len(token) != 32 {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token size")
	}
	btoken, err := hex.DecodeString(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token format")
	}
	_ = btoken
	return echo.NewHTTPError(http.StatusNotImplemented)
}
