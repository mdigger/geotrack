package main

import (
	"net/http"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
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
