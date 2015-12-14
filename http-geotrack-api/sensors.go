package main

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	"github.com/mdigger/geotrack/sensors"
	"gopkg.in/mgo.v2"
)

const serviceNameSensors = "device.sensor"

// getSensors отдает список изменений сенсоров устройства.
func getSensors(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	limit, err := strconv.ParseUint(c.Query("limit"), 10, 16)
	if err != nil || limit < 1 {
		limit = listLimit
	}
	lastID := c.Query("last")
	// запрашиваем список устройств постранично
	sensors, err := sensorsDB.Get(groupID, deviceID, int(limit), lastID)
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("tracksDB error: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, sensors)
}

// postSensors добавляет новые данные о сенсорах устройства в хранилище.
func postSensors(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	deviceID := c.Param("device-id")
	var sensors = make([]sensors.SensorData, 0)
	err := c.Bind(&sensors)
	if err != nil || len(deviceID) < 12 {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	// добавляем идентификатор группы и устройства
	for i, sensor := range sensors {
		sensor.DeviceID = deviceID
		sensor.GroupID = groupID
		sensors[i] = sensor
	}
	// пропускаем через NATS, а не на прямую в базу
	err = nce.Publish(serviceNameSensors, sensors)
	// err = sensorsDB.Add(sensors...)
	if err != nil {
		llog.Error("sensors NATS publishing error: %v", err)
		return err
	}
	return c.NoContent(http.StatusOK)
}
