package main

import (
	"net/http"

	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// login читает заголовок запроса с HTTP Basic авторизацией, проверяет пользователя
// по базе данных и отдает в ответ авторизационный ключ в формате JWT.
func login(c *echo.Context) error {
	// получаем пароль из заголовка HTTP Basic авторизации
	username, password, ok := c.Request().BasicAuth()
	if !ok {
		c.Response().Header().Set(echo.WWWAuthenticate, "Basic realm=Restricted")
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	// получаем из хранилища информацию о пользователе
	user, err := usersDB.Get(username)
	if err == mgo.ErrNotFound {
		return echo.NewHTTPError(http.StatusForbidden)
	}
	if err != nil {
		llog.Error("userDB error: %v", err)
		return err
	}
	// сравниваем сохраненный пароль с тем, что указали в заголовке
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return echo.NewHTTPError(http.StatusForbidden)
	}
	// генерируем JWT-токен
	tokenString, err := tokenEngine.Token(map[string]interface{}{
		"id":    user.ID.Hex(),
		"group": user.GroupID,
	})
	if err != nil {
		llog.Error("tokenEngine error: %v", err)
		return err
	}
	// отдаем в ответ сервера
	response := c.Response()
	response.Header().Set(echo.ContentType, "application/jwt")
	response.WriteHeader(http.StatusOK)
	response.Write([]byte(tokenString))
	return nil
}

// auth является вспомогательной функцией, проверяющей и разбирающей токен с авторизационной
// информацией в HTTP-заголовке. Разобранная информация сохраняется в контексте запроса.
func auth(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req := c.Request()                                  // получаем доступ к HTTP-запросу
		if req.Header.Get(echo.Upgrade) == echo.WebSocket { // пропускаем запросы WebSocket
			return nil
		}
		data, err := tokenEngine.ParseRequest(req) // разбираем токен из запроса
		if err != nil {
			llog.Warn("Bad token: %v", err)
			return echo.NewHTTPError(http.StatusForbidden)
		}
		groupID := data["group"].(string)
		userID := data["id"].(string)
		if !bson.IsObjectIdHex(userID) {
			llog.Warn("Bad user Object ID: %v", userID)
			return echo.NewHTTPError(http.StatusForbidden)
		}
		userObjectId := bson.ObjectIdHex(userID)
		// проверяем, что пользователь есть и входит в эту группу
		exists, err := usersDB.Check(groupID, userObjectId)
		if err != nil {
			llog.Error("usersDB error: %v", err)
			return err
		}
		if !exists {
			llog.Debug("Auth not exist: %v (%v)", userID, groupID)
			return echo.NewHTTPError(http.StatusForbidden)
		}
		c.Set("GroupID", groupID) // сохраняем данные в контексте запроса
		c.Set("ID", userObjectId)
		llog.Debug("Auth: %v (%v)", userID, groupID)
		return h(c) // выполняем основной обработчик
	}
}
