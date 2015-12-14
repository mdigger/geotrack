package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/mdigger/geotrack/tracks"
	"gopkg.in/mgo.v2"
)

const (
	listLimit         = 200 // лимит при отдаче списка треков
	serviceNameTracks = "device.track"
)

// getTracks отдает всю историю с координатами трекинга браслета, разбивая ее на порции.
func getTracks(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	limit, err := strconv.ParseUint(c.Query("limit"), 10, 16)
	if err != nil || limit < 1 {
		limit = listLimit
	}
	lastID := c.Query("last")
	// запрашиваем список устройств постранично
	tracks, err := tracksDB.Get(groupID, deviceID, int(limit), lastID)
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("tracksDB error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, tracks)
}

// postTracks добавляет новые данные треков устройства в хранилище.
func postTracks(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	var tracks = make([]tracks.TrackData, 0)
	err := c.Bind(&tracks)
	if err != nil || len(deviceID) < 12 {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	// добавляем идентификатор группы и устройства
	for i, track := range tracks {
		track.DeviceID = deviceID
		track.GroupID = groupID
		tracks[i] = track
	}
	// пропускаем через NATS, а не на прямую в базу
	err = nce.Publish(serviceNameTracks, tracks)
	// err = tracksDB.Add(tracks...)
	if err != nil {
		llog.Error("tracks NATS publishing error: %v", err)
		return err
	}
	return c.NoContent(http.StatusOK)
}
