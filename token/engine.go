package token

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// AccessTokenParamName описывает название параметра с токеном в запросе, если токен передается
// не в заголовке. Если в качестве значения задать пустую строку, то система перестанет
// поддерживать возможность передачи токена в виде параметра.
var AccessTokenParamName = "token"

// Engine описывает класс для работы с токенами в формате JSON Web Token.
type Engine struct {
	issuer    string        // название сервиса
	expire    time.Duration // время жизни ключа
	cryptoKey []byte        // ключ для подписи JWT
}

// Init инициализирует и возвращает класс для работы с токенами.
// Если ключ для подписи указан пустой, то формируется новый случайный ключ.
func Init(issuer string, expire time.Duration, cryptoKey []byte) (*Engine, error) {
	if cryptoKey == nil {
		cryptoKey = make([]byte, 256)
		if _, err := rand.Read(cryptoKey); err != nil {
			return nil, err
		}
	}
	return &Engine{
		issuer:    issuer,
		expire:    expire,
		cryptoKey: cryptoKey,
	}, nil
}

// Token формирует и возвращает токен в формате JWT.
func (e *Engine) Token(items map[string]interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256) // генерируем новый токен
	for key, value := range items {          // добавляем в него наши данные
		token.Claims[key] = value
	}
	if e.issuer != "" { // добавляем информацию о сервисе
		token.Claims["iss"] = e.issuer
	}
	if e.expire != 0 { // время жизни токена
		token.Claims["exp"] = time.Now().Add(e.expire).Unix()
	}
	return token.SignedString(e.cryptoKey)
}

// verify является функцией для проверки целостности токена.
func (e *Engine) verify(token *jwt.Token) (key interface{}, err error) {
	key = e.cryptoKey // ключ, используемый для подписи
	// проверяем метод вычисления сигнатуры и обязательные поля
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		err = fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	} else if e.issuer != "" && token.Claims["iss"] != e.issuer {
		err = fmt.Errorf("unexpected Issuer: %v", token.Claims["iss"])
	} else if _, ok := token.Claims["exp"]; e.expire != 0 && !ok {
		err = errors.New("missing Expire")
	}
	return
}

// Parse разбирает токен, проверяет его валидность и возвращает данные из него.
func (e *Engine) Parse(tokenString string) (data map[string]interface{}, err error) {
	token, err := jwt.Parse(tokenString, e.verify)
	if err != nil {
		return nil, err
	}
	return token.Claims, nil
}

// ParseRequest разбирает токен из HTTP-запроса. Токен может быть передан как в заголовке
// запроса авторизации, с типом авторизации "Bearer", так и в параметре или поле формы
// с имененен, определеннов в глобальной переменной AccessTokenParamName.
func (e *Engine) ParseRequest(req *http.Request) (data map[string]interface{}, err error) {
	if ah := req.Header.Get("Authorization"); ah != "" {
		if len(ah) > 6 && strings.ToUpper(ah[0:6]) == "BEARER" {
			return e.Parse(ah[7:])
		}
	}
	if AccessTokenParamName != "" {
		if tokStr := req.FormValue(AccessTokenParamName); tokStr != "" {
			return e.Parse(tokStr)
		}
	}
	return nil, jwt.ErrNoTokenInRequest
}

// CryptoKey возвращает ключ, используемый для подписи, в виде строки (base64-encoded).
func (e *Engine) CryptoKey() string {
	return base64.StdEncoding.EncodeToString(e.cryptoKey)
}
