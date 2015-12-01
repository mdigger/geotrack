package main

import (
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/mdigger/geotrack/mongo"
	"github.com/mdigger/geotrack/tracks"
)

var (
	db *tracks.DB
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
	flag.Parse()

	mdb, err := mongo.Connect(*mongoURL)
	if err != nil {
		log.Println("Error connecting to MongoDB:", err)
		return
	}
	defer mdb.Close()

	// инициализируем хранилище с информацией о треках
	db, err = tracks.InitDB(mdb)
	if err != nil {
		log.Println("Error initializing TrackDB:", err)
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
	e.Run(*addr)
}

func index(c *echo.Context) error {
	// получаем список устройств
	deviceids, err := db.GetDevicesID()
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "index.html", deviceids)
}

func current(c *echo.Context) error {
	deviceID := c.Param("deviceid")
	track, err := db.GetLast(deviceID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "current.html", track)
}

func history(c *echo.Context) error {
	deviceID := c.Param("deviceid")
	tracks, err := db.GetDay(deviceID)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "history.html", tracks)
}
