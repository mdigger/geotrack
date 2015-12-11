package parser

import (
	"encoding/hex"
	"encoding/json"
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
func ParseWiFi(s string) ([]geolocate.WifiAccessPoint, []string, error) {
	wifis := make([]WiFiData, 0)
	if err := json.Unmarshal([]byte(s), &wifis); err != nil {
		return nil, nil, err
	}
	names := make([]string, len(wifis))
	for i, item := range wifis {
		names[i] = item.Name
	}
	wifiAccessPoints := make([]geolocate.WifiAccessPoint, len(wifis))
	for i, wifi := range wifis {
		signalStrength, err := strconv.ParseUint(wifi.SignalStrength, 16, 16)
		if err != nil {
			return nil, nil, fmt.Errorf("bad WiFi signal strength: %s", wifi.SignalStrength)
		}
		mac, err := hex.DecodeString(wifi.MacAddress)
		if err != nil || len(mac) != 6 {
			return nil, nil, fmt.Errorf("bad WiFi mac address: %s", wifi.MacAddress)
		}
		wifiAccessPoints[i] = geolocate.WifiAccessPoint{
			MacAddress:     fmt.Sprintf("%X:%X:%X:%X:%X:%X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]),
			SignalStrength: int16(math.Log10(float64(signalStrength)/1000) * 100),
		}
	}
	return wifiAccessPoints, names, nil
}
