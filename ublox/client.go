package ublox

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/mdigger/geotrack/geo"
)

var (
	// RequestTimeout описывает время ожидания от сервера, которое используется при инициализации
	// клиента.
	RequestTimeout = time.Second * 5
	// Servers описывает список серверов для запросов данных.
	Servers = []string{
		"http://online-live1.services.u-blox.com/GetOnlineData.ashx",
		"http://online-live2.services.u-blox.com/GetOnlineData.ashx",
	}
	Pacc = 100000 // расстояние погрешности в метрах
)

// Client описxывает сервис получения данных.
type Client struct {
	token  string       // The authorization token supplied by u-blox when a client registers to use the service
	client *http.Client // HTTP-клиент
}

// NewClient возвращает новый инициализированный провайдер для получения данных.
func NewClient(token string) *Client {
	return &Client{
		token: token,
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// GetOnline запрашивает сервер u-blox и получает данные для указанной точки и профиля устройства.
func (c *Client) GetOnline(point geo.Point, profile Profile) ([]byte, error) {
	// формируем строку запроса
	var query = new(bytes.Buffer)
	query.WriteString("token=")
	query.WriteString(c.token)
	if profile.Format != "" {
		fmt.Fprintf(query, ";format=%s", profile.Format)
	}
	if len(profile.Datatype) > 0 {
		fmt.Fprintf(query, ";datatype=%s", strings.Join(profile.Datatype, ","))
	}
	if len(profile.GNSS) > 0 {
		fmt.Fprintf(query, ";gnss=%s", strings.Join(profile.GNSS, ","))
	}
	if !point.IsZero() {
		fmt.Fprintf(query, ";lon=%f;lat=%f", point.Longitude(), point.Latitude())
		if Pacc >= 0 && Pacc != 300000 && Pacc < 6000000 {
			fmt.Fprintf(query, ";pacc=%d", Pacc)
		}
		if profile.FilterOnPos {
			query.WriteString(";filteronpos")
		}
	}

	var n = 0 // номер сервера для запроса из списка
repeatOnTimeout:
	// формируем URL запроса
	reqURL := fmt.Sprintf("%s?%s", Servers[n], query.String())
	log.Println("UBLOX:", reqURL)     // выводим в лог URL запроса
	resp, err := c.client.Get(reqURL) // осуществляем запрос к серверу на получение данных
	if err != nil {
		// проверяем, что ошибка таймаута получения данных
		if e, ok := err.(net.Error); ok && e.Timeout() {
			if len(Servers) > n+1 {
				n++                  // увеличиваем номер используемого сервера из списка
				goto repeatOnTimeout // повторяем запрос с новым сервером
			}
		}
		return nil, err
	}
	defer resp.Body.Close()
	// TODO: нужно ли проверять HTTP-коды ответов
	log.Printf("UBLOX: %s [%d bytes]", resp.Status, resp.ContentLength)
	// читаем и возвращаем данные ответа
	return ioutil.ReadAll(resp.Body)
}
