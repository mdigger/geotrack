package ublox

import (
	"fmt"
	"testing"

	"github.com/mdigger/geotrack/geo"
)

var (
	token     = "I6KKO4RU_U2DclBM9GVyrA"
	pointWork = geo.NewPoint(37.57351, 55.715084)  // работа
	pointHome = geo.NewPoint(37.589248, 55.765944) // дом
)

func TestClient(t *testing.T) {
	ubox := NewClient(token)
	data, err := ubox.GetOnline(pointWork, DefaultProfile)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
