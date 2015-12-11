package parser

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"

	"github.com/mdigger/geolocate"
)

type WiFiData struct {
	MacAddress     string `json:"m"`
	Name           string `json:"i"`
	SignalStrength string `json:"s"`
}

// ParseWiFi разбирает и возвращает список с информацией о WiFi-станциях.
func ParseWiFi(wifis ...WiFiData) ([]geolocate.WifiAccessPoint, error) {
	wifiAccessPoints := make([]geolocate.WifiAccessPoint, len(wifis))
	for i, wifi := range wifis {
		signalStrength, err := strconv.ParseUint(wifi.SignalStrength, 16, 16)
		if err != nil {
			return nil, fmt.Errorf("bad WiFi signal strength: %s", wifi.SignalStrength)
		}
		mac, err := hex.DecodeString(wifi.MacAddress)
		if err != nil || len(mac) != 6 {
			return nil, fmt.Errorf("bad WiFi mac address: %s", wifi.MacAddress)
		}
		wifiAccessPoints[i] = geolocate.WifiAccessPoint{
			MacAddress:     fmt.Sprintf("%X:%X:%X:%X:%X:%X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]),
			SignalStrength: int16(math.Log10(float64(signalStrength)/1000) * 100),
		}
	}
	return wifiAccessPoints, nil
}
