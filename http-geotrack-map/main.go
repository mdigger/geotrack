package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/places"
	"github.com/mdigger/geotrack/tracks"
	"github.com/mdigger/geotrack/users"
	"gopkg.in/mgo.v2"
)

var (
	tracksDB *tracks.DB
	placesDB *places.DB
	groupID  = users.SampleGroupID
)

// Template provides HTML template rendering
type Template struct {
	templates *template.Template
}

// Render HTML
func (t *Template) Render(w io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	mongoURL := flag.String("mongodb", "mongodb://localhost/watch", "MongoDB connection URL")
	addr := flag.String("http", ":8080", "Server address & port")
	docker := flag.Bool("docker", false, "for docker")
	flag.Parse()

	// Если запускается внутри контейнера
	if *docker {
		tmp := os.Getenv("MONGODB")
		mongoURL = &tmp
	}

	mdb, err := mongo.Connect(*mongoURL)
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	// инициализируем хранилище с информацией о треках
	tracksDB, err = tracks.InitDB(mdb)
	if err != nil {
		log.Println("Error initializing TrackDB:", err)
		return
	}
	// инициализируем хранилище с информацией о местах
	placesDB, err = places.InitDB(mdb)
	if err != nil {
		log.Println("Error initializing PlaceDB:", err)
		return
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.SetRenderer(&Template{templates: template.Must(
		template.ParseFiles("index.html", "current.html", "history.html"))})
	e.Get("/", index)
	e.Get("/:deviceid", current)
	e.Get("/:deviceid/history", history)
	e.ServeFile("/edit", "placeeditor.html")
	e.Run(*addr)
}

func index(c *echo.Context) error {
	deviceids, err := tracksDB.GetDevicesID(groupID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "index.html", deviceids)
}

func current(c *echo.Context) error {
	deviceID := c.Param("deviceid")
	track, err := tracksDB.GetLast(groupID, deviceID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}
	return c.Render(http.StatusOK, "current.html", track)
}

func history(c *echo.Context) error {
	deviceID := c.Param("deviceid")
	dayTracks, err := tracksDB.GetDay(groupID, deviceID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}
	groupPlaces, err := placesDB.GetAll(groupID)
	if err != nil && err != mgo.ErrNotFound {
		return err
	}
	data := struct {
		Tracks []tracks.Track
		Places []places.Place
	}{
		Tracks: dayTracks,
		Places: groupPlaces,
	}
	return c.Render(http.StatusOK, "history.html", &data)
}
