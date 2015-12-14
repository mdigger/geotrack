package main

import (
	"encoding/hex"
	"net/http"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

// postRegister регистрирует устройство, чтобы на него можно было отсылать push-уведомления.
func postRegister(c *echo.Context) error {
	pushType := c.Param("push-type") // получаем идентификатор типа уведомлений
	switch pushType {
	case "apns": // Apple Push Notification
	case "gcm": // Google Cloud Messages
	default: // не поддерживаемы тип
		return echo.NewHTTPError(http.StatusNotFound)
	}
	groupID := c.Get("GroupID").(string)  // идентификатор группы
	userID := c.Get("ID").(bson.ObjectId) // идентификатор пользователя
	token := c.Form("token")              // токен устройства
	if len(token) != 32 {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token size")
	}
	btoken, err := hex.DecodeString(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token format")
	}
	_, _, _ = btoken, groupID, userID
	return echo.NewHTTPError(http.StatusNotImplemented)
}

// deleteRegister регистрирует устройство, чтобы на него можно было отсылать push-уведомления.
func deleteRegister(c *echo.Context) error {
	pushType := c.Param("push-type") // получаем идентификатор типа уведомлений
	switch pushType {
	case "apns": // Apple Push Notification
	case "gcm": // Google Cloud Messages
	default: // не поддерживаемы тип
		return echo.NewHTTPError(http.StatusNotFound)
	}
	groupID := c.Get("GroupID").(string)  // идентификатор группы
	userID := c.Get("ID").(bson.ObjectId) // идентификатор пользователя
	token := c.Param("token")             // токен устройства
	if len(token) != 32 {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token size")
	}
	btoken, err := hex.DecodeString(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad token format")
	}
	_, _, _ = btoken, groupID, userID
	return echo.NewHTTPError(http.StatusNotImplemented)
}
