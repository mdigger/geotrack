package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
)

const (
	serviceNamePairingKey = "device.pair.key"
	natsRequestTimeout    = time.Second * 10
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

func postDevicePairing(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	_, _ = groupID, deviceID
	var pairingKey struct {
		Key string
	}
	err := c.Bind(&pairingKey) // читаем ключ из запроса
	if err != nil || len(pairingKey.Key) < 4 {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	var deviceIDResp string
	err = nce.Request(serviceNamePairingKey, pairingKey.Key, &deviceIDResp, natsRequestTimeout)
	if err != nil {
		llog.Error("NATS Pairing Key response error: %v", err)
		return err
	}
	if deviceIDResp == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if deviceIDResp == deviceID {
		// TODO: реально связать в базе
		return echo.NewHTTPError(http.StatusOK)
	}
	if deviceID == "" {
		return c.JSON(http.StatusOK, map[string]string{"ID": deviceIDResp})
	}
	return echo.NewHTTPError(http.StatusBadRequest)
}
