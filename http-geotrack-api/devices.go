package main

import (
	"net/http"

	"gopkg.in/mgo.v2"

	"github.com/labstack/echo"
)

const (
	tracksLimit = 200 // лимит при отдаче списка треков
)

// getDevices отдает список зарегистрированных устройств, которые относятся к той же
// группе, что и текущий пользователь.
func getDevices(c *echo.Context) error {
	// TODO: возвращать все устройства, а не только те, треки по которым сохранились
	groupID := c.Get("GroupID").(string)             // получаем идентификатор группы
	deviceIDs, err := tracksDB.GetDevicesID(groupID) // запрашиваем список устройств
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("tracksDB error: %v", err)
		return err
	}
	return c.JSON(http.StatusOK, deviceIDs)
}

// getDeviceCurrent отдает последние данные с координатами браслета.
func getDeviceCurrent(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)              // получаем идентификатор группы
	deviceID := c.Param("device-id")                  // получаем идентификатор устройства
	track, err := tracksDB.GetLast(groupID, deviceID) // запрашиваем список устройств
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("tracksDB error: %v", err)
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
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("tracksDB error: %v", err)
		return err
	}
	return c.JSON(http.StatusOK, tracks)
}
