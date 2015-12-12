package main

import (
	"net/http"

	"gopkg.in/mgo.v2"

	"github.com/labstack/echo"
	"github.com/mdigger/geotrack/tracks"
)

const (
	tracksLimit = 200 // лимит при отдаче списка треков
)

// getDevices отдает список зарегистрированных устройств, которые относятся к той же
// группе, что и текущий пользователь.
func getDevices(c *echo.Context) error {
	// TODO: возвращать все устройства, а не только те, треки по которым сохранились
	groupID := c.Get("GroupID").(string)
	deviceIDs, err := tracksDB.GetDevicesID(groupID)
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
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	track, err := tracksDB.GetLast(groupID, deviceID)
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
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	lastID := c.Query("last")
	// запрашиваем список устройств постранично
	tracks, err := tracksDB.Get(groupID, deviceID, tracksLimit, lastID)
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("tracksDB error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, tracks)
}

// postDeviceHistory добавляет новые данные треков устройства в хранилище.
func postDeviceHistory(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	var tracks = make([]tracks.TrackData, 0)
	err := c.Bind(&tracks)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	// добавляем идентификатор группы и устройства
	for i, track := range tracks {
		track.DeviceID = deviceID
		track.GroupID = groupID
		tracks[i] = track
	}
	err = tracksDB.Add(tracks...)
	if err != nil {
		llog.Error("tracksDB error: %v", err)
		return err
	}
	return c.NoContent(http.StatusOK)
}
