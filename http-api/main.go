package main

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	logger "github.com/labstack/gommon/log"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/places"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/users"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
)

var (
	usersDB  *users.DB             // хранилище пользователей
	placesDB *places.DB            // хранилище мест
	tracksDB *tracks.DB            // хранилище треков
	groupID  = users.SampleGroupID // уникальный идентификатор группы
	// jwtCryptoKey используется подписи JWT
	jwtCryptoKey = []byte(`очень секьюрная строка, используемая для подписи`)
	log          *logger.Logger // вывод информации в лог
)

const (
	jwtExpireDuration = time.Minute * 30    // время жизни JWT-токена
	jwtIssuer         = "com.xyzrd.tracker" // идентификатор издателя
	jwtSubject        = "user"              // тип информации
	tracksLimit       = 200                 // лимит при отдаче списка треков
)

func main() {
	fmt.Println(base64.StdEncoding.EncodeToString(jwtCryptoKey))
	addr := flag.String("http", ":8080", "Server address & port")
	mongoURL := flag.String("mongodb", "mongodb://localhost/watch", "MongoDB connection URL")
	flag.Parse()

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
	// TODO: убрать по мере имплементации
	_ = usersDB
	_ = placesDB
	_ = tracksDB
	groupID = usersDB.GetSampleGroupID() // временная инициализация пользователей
	// TODO: добавить случайную генерацию ключа при каждом запуске сервера

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	apiV1 := e.Group("/api/v1")                                   // группа URL для обработки API версии 1.0.
	apiV1.Get("/login", login)                                    // аторизация пользователя
	apiV1Sec := apiV1.Group("")                                   // группа запросов с авторизацией
	apiV1Sec.Use(jwtAuth)                                         // добавляем проверку токена в заголовке
	apiV1Sec.Get("/users", getUserslList)                         // возвращает список пользователей
	apiV1Sec.Get("/places", getPlaceslList)                       // возвращает список интересующих мест
	apiV1Sec.Get("/devices", getDeviceslList)                     // возвращает список устройств
	apiV1Sec.Get("/devices/:device-id", getDeviceCurrent)         // возвращает последнюю точку трекинга устройства
	apiV1Sec.Get("/devices/:device-id/history", getDeviceHistory) // возвращает список треков устройства
	apiV1Sec.Post("/register", postRegister)                      // регистрирует устройство для отправки push-сообщений

	fmt.Println(e.URI(getUserslList))

	log.Info("Starting HTTP server at %q...", *addr)
	e.Run(*addr)
}

// login читает заголовок запроса с HTTP Basic авторизацией, проверяет пользователя
// по базе данных и отдает в ответ авторизационный ключ в формате JWT.
func login(c *echo.Context) error {
	log.Trace("->login")
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
	token := jwt.New(jwt.SigningMethodHS256)
	if jwtIssuer != "" {
		token.Claims["iss"] = jwtIssuer
	}
	token.Claims["sub"] = "user"
	token.Claims["exp"] = time.Now().Add(jwtExpireDuration).Unix()
	token.Claims["id"] = user.ID
	token.Claims["group"] = user.GroupID
	token.Claims["name"] = user.Name
	token.Claims["icon"] = user.Icon
	tokenString, err := token.SignedString(jwtCryptoKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	// отдаем в ответ сервера
	response := c.Response()
	response.Header().Set(echo.ContentType, "application/jwt")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(tokenString))
	return nil
}

// jwtAuth является вспомогательной функцией, проверяющей и разбирающей авторизационную информацию
// в заголовке в формате JWT. Разобранная информация сохраняется в контексте под именем "User".
func jwtAuth(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		log.Trace("~>jwtAuth")
		header := c.Request().Header
		// пропускаем запросы WebSocket
		if header.Get(echo.Upgrade) == echo.WebSocket {
			return nil
		}
		// получаем заголовок авторизации ипроверяем, что авторизация с JWT-токеном
		auth := header.Get("Authorization")
		if len(auth) < 7 || strings.ToUpper(auth[0:6]) != "BEARER" {
			log.Debug("Not BEARER:\n%v", auth)
			return echo.NewHTTPError(http.StatusForbidden)
		}
		// разбираем и проверяем сам токен
		token, err := jwt.Parse(auth[7:], func(token *jwt.Token) (key interface{}, err error) {
			// проверяем метод вычисления сигнатуры
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				err = fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			} else if token.Claims["iss"] != jwtIssuer {
				err = fmt.Errorf("Unexpected Issuer: %v", token.Claims["iss"])
			} else if token.Claims["sub"] != jwtSubject {
				err = fmt.Errorf("Unexpected Subject: %v", token.Claims["sub"])
			} else if token.Claims["id"] == "" {
				err = errors.New("Unexpected User ID")
			} else if token.Claims["group"] == "" {
				err = errors.New("Unexpected Group ID")
			}
			key = jwtCryptoKey // ключ, используемый для подписи
			return
		})
		// возвращаем ошибку, если токен не валиден
		if err != nil || !token.Valid {
			log.Debug("Bad JWT-token: %v", err)
			return echo.NewHTTPError(http.StatusForbidden)
		}
		// сохраняем данные в контексте запроса
		c.Set("ID", token.Claims["id"])
		c.Set("GroupID", token.Claims["group"])
		// TODO: проверить, что пользователь есть и не заблокирован.
		// выполняем основной обработчик
		return h(c)
	}
}

// getUserslList отдает список зарегистрированных пользователей, которые относятся к той же
// группе, что и текущий пользователь.
func getUserslList(c *echo.Context) error {
	log.Trace("->getUserslList")
	groupID := c.Get("GroupID").(string)    // получаем идентификатор группы
	users, err := usersDB.GetUsers(groupID) // запрашиваем список пользователей
	if err != nil {
		return err
	}
	return c.JSONIndent(http.StatusOK, users, "", "\t")
}

// getPlaceslList список мест, зарегистрированны для группы пользователей
func getPlaceslList(c *echo.Context) error {
	log.Trace("->getPlaceslList")
	groupID := c.Get("GroupID").(string) // получаем идентификатор группы
	places, err := placesDB.Get(groupID) // запрашиваем список мест
	if err != nil {
		return err
	}
	return c.JSONIndent(http.StatusOK, places, "", "\t")
}

// getDeviceslList отдает список зарегистрированных устройств, которые относятся к той же
// группе, что и текущий пользователь.
func getDeviceslList(c *echo.Context) error {
	log.Trace("->getDeviceslList")
	// TODO: возвращать все устройства, а не только те, треки по которым сохранились
	groupID := c.Get("GroupID").(string)             // получаем идентификатор группы
	deviceIDs, err := tracksDB.GetDevicesID(groupID) // запрашиваем список устройств
	if err != nil {
		return err
	}
	return c.JSONIndent(http.StatusOK, deviceIDs, "", "\t")
}

// getDeviceCurrent отдает последние данные с координатами браслета.
func getDeviceCurrent(c *echo.Context) error {
	log.Trace("->getDeviceCurrent")
	groupID := c.Get("GroupID").(string)              // получаем идентификатор группы
	deviceID := c.Param("device-id")                  // получаем идентификатор устройства
	track, err := tracksDB.GetLast(groupID, deviceID) // запрашиваем список устройств
	if err != nil {
		return err
	}
	return c.JSONIndent(http.StatusOK, track, "", "\t")
}

// getDeviceHistory отдает всю историю с координатами трекинга браслета, разбивая ее на порции.
func getDeviceHistory(c *echo.Context) error {
	log.Trace("->getDeviceHistory")
	groupID := c.Get("GroupID").(string) // получаем идентификатор группы
	deviceID := c.Param("device-id")     // получаем идентификатор устройства
	lastID := c.Query("last")            // получаем идентификатор последнего полученного трека
	// запрашиваем список устройств по странично
	tracks, err := tracksDB.Get(groupID, deviceID, tracksLimit, lastID)
	if err != nil {
		return err
	}
	return c.JSONIndent(http.StatusOK, tracks, "", "\t")
}

// postRegister регистрирует устройство, чтобы на него можно было отсылать push-уведомления.
func postRegister(c *echo.Context) error {
	log.Trace("->postRegister")
	return echo.NewHTTPError(http.StatusNotImplemented)
}
