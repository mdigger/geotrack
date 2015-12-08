package main

import (
	"net"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/color"
)

func Logger() echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			res := c.Response()
			logger := c.Echo().Logger()

			remoteAddr := req.RemoteAddr
			if ip := req.Header.Get(echo.XRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header.Get(echo.XForwardedFor); ip != "" {
				remoteAddr = ip
			} else {
				remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
			}

			start := time.Now()
			if err := h(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			method := req.Method
			path := req.URL.Path
			if path == "" {
				path = "/"
			}
			size := res.Size()

			n := res.Status()
			code := color.Green(n)
			switch {
			case n >= 500:
				code = color.Red(n)
			case n >= 400:
				code = color.Yellow(n)
			case n >= 300:
				code = color.Cyan(n)
			}
			logger.Info("%14v %-20s %3s %4s %s [%d]",
				stop.Sub(start), // Продолжительность обработки запроса
				remoteAddr,      // IP-адрес того, кто осуществляет запрос
				code,            // Статус ответа
				method,          // Метод HTTP-запроса
				path,            // Запрашиваемый URL
				size,            // размер ответа
			)
			return nil
		}
	}
}
