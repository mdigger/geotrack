package main

import (
	"net/http"

	"github.com/labstack/echo"
	"gopkg.in/mgo.v2"
)

// getUsers отдает список зарегистрированных пользователей, которые относятся к той же
// группе, что и текущий пользователь.
func getUsers(c *echo.Context) error {
	groupID := c.Get("GroupID").(string)    // получаем идентификатор группы
	users, err := usersDB.GetUsers(groupID) // запрашиваем список пользователей
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	if err != nil {
		llog.Error("usersDB error: %v", err)
		return err
	}
	return c.JSON(http.StatusOK, users)
}
