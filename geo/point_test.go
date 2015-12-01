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
	const rad = 6372795
	// координаты двух точек
	p1 := NewPoint(-139.398, 77.1539)
	p2 := NewPoint(-139.55, -77.1804)
	// в радианах
	lat1 := p1.Latitude() * math.Pi / 180
	lat2 := p2.Latitude() * math.Pi / 180
	long1 := p1.Longitude() * math.Pi / 180
	long2 := p2.Longitude() * math.Pi / 180
	// косинусы и синусы широт и разницы долгот
	cl1 := math.Cos(lat1)
	cl2 := math.Cos(lat2)
	sl1 := math.Sin(lat1)
	sl2 := math.Sin(lat2)
	delta := long2 - long1
	cdelta := math.Cos(delta)
	sdelta := math.Sin(delta)
	// вычисления длины большого круга
	y := math.Sqrt(math.Pow(cl2*sdelta, 2) + math.Pow(cl1*sl2-sl1*cl2*cdelta, 2))
	x := sl1*sl2 + cl1*cl2*cdelta
	ad := math.Atan2(y, x)
	dist := ad * rad
	fmt.Printf("Dist: %f\n", dist)
	// вычисление начального азимута
	// x = (cl1 * sl2) - (sl1 * cl2 * cdelta)
	// y = sdelta * cl2
	// z := (math.Atan(-y / x)) / (math.Pi / 180)
	// if x < 0 {
	// 	z = z + 180.
	// }
	// z2 := (z+180)%360 - 180
	// z2 = -z2 * math.Pi / 180
	// anglerad2 := z2 - ((2 * math.Pi) * math.Floor((z2 / (2 * math.Pi))))
	// angledeg := (anglerad2 * 180) / math.Pi
}
