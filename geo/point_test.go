package geo

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
)

func TestPoint(t *testing.T) {
	point := NewPoint(37.57351, 55.715084)
	data, err := json.Marshal(point)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	var point2 Point
	err = json.Unmarshal(data, &point2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(point2)
	fmt.Println(point.Distance(NewPoint(37.589248, 55.765944)))

	p1 := NewPoint(77.1539, -139.398)
	p2 := NewPoint(-77.1804, -139.55)
	fmt.Printf("%f\n", p1.Distance(p2))
}

func TestGeoPoint(t *testing.T) {
	point := NewPoint(37.57351, 55.715084)
	fmt.Println(point.Geo())
	// pretty.Println(point.GeoPolygon(500))
}

func TestPoint3(t *testing.T) {
	const earthRadius float64 = 6378137.0
	p := NewPoint(37.57351, 55.715084)
	var radius float64 = 500.0
	centerLat := p.Latitude()
	centerLng := p.Longitude()
	dLat := radius / (earthRadius * math.Pi / 180.0)
	dLng := radius / (earthRadius * math.Cos(math.Pi/180.0*centerLat) * math.Pi / 180.0)
	segments := 16
	dRad := 2.0 * math.Pi / float64(segments)
	for i := 0; i <= segments; i++ {
		rads := dRad * float64(i)
		y := math.Sin(rads)
		x := math.Cos(rads)
		if math.Abs(y) < 0.01 {
			y = 0.0
		}
		if math.Abs(x) < 0.01 {
			x = 0.0
		}
		fmt.Printf("[%f,%f],\n", centerLng+y*dLng, centerLat+x*dLat)
	}
}

func TestDistance(t *testing.T) {
	p1 := NewPoint(37.57351, 55.715084)
	p2 := NewPoint(37.589248, 55.765944)
	fmt.Printf("Dist: %f\n", p1.Distance(p2))
	for i := 0; i < 16; i++ {
		fmt.Printf("myMap.geoObjects.add(new ymaps.Placemark(%v));\n", p1.Move(500.0, float64(i)*360.0/16.0))
	}
}
