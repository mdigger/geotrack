package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/mdigger/geotrack/places"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// getPlaces список мест, зарегистрированны для группы пользователей.
func getPlaces(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	places, err := placesDB.GetAll(groupID)
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("placesDB error: %v", err)
		return err
	}
	return c.JSON(http.StatusOK, places)
}

// getPlace возвращает информацию о месте с заданным идентификатором.
func getPlace(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	placeID := c.Query("place-id")
	if !bson.IsObjectIdHex(placeID) {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	place, err := placesDB.Get(groupID, bson.ObjectIdHex(placeID))
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("placesDB error: %v", err)
		return err
	}
	return c.JSON(http.StatusOK, place)
}

// postPlace добавляет новое определение места.
func postPlace(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	var place places.Place // описание места
	err := c.Bind(&place)  // разбираем описание места из запроса
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if place.Circle == nil && place.Polygon == nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	place.ID = ""
	place.GroupID = groupID
	id, err := placesDB.Save(place)
	if err != nil {
		llog.Error("placesDB error: %v", err)
		return err
	}
	c.Response().Header().Set("Location", e.URL(getPlace, id.Hex()))
	return c.JSON(http.StatusCreated, map[string]interface{}{"ID": id.Hex()})
}

// putPlace изменяет определение уже существующего места.
func putPlace(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	placeID := c.Query("place-id")
	if !bson.IsObjectIdHex(placeID) {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	var place places.Place // описание места
	err := c.Bind(&place)  // разбираем описание места из запроса
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if place.Circle == nil && place.Polygon == nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	place.ID = bson.ObjectIdHex(placeID)
	place.GroupID = groupID
	_, err = placesDB.Save(place)
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("placesDB error: %v", err)
		return err
	}
	return c.NoContent(http.StatusOK)
}

// deletePlace удаляет определение места.
func deletePlace(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)
	placeID := c.Query("place-id")
	if !bson.IsObjectIdHex(placeID) {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	err := placesDB.Delete(groupID, bson.ObjectIdHex(placeID))
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("placesDB error: %v", err)
		return err
	}
	return c.NoContent(http.StatusOK)
}
