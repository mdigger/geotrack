package geo

import (
	"fmt"
	"math"
)

// Point описывает координаты точки.
type Point [2]float64

// NewPoint возвращает новое описание точки с указанными координатами.
func NewPoint(lon, lat float64) *Point {
	if lon < -180 || lon > 180 {
		panic("bad longitude")
	}
	if lat < -90 || lat > 90 {
		panic("bad latitude")
	}
	return &Point{lon, lat}
}

// Longitude возвращает долготу точки.
func (p *Point) Longitude() float64 {
	if p == nil {
		return 0
	}
	return p[0]
}

// Latitude возвращает широту точки.
func (p *Point) Latitude() float64 {
	if p == nil {
		return 0
	}
	return p[1]
}

// String возвращает строковое представление координат точки.
func (p *Point) String() string {
	return fmt.Sprintf("[%f,%f]", p.Longitude(), p.Latitude())
}

// GeoPoint возвращает представление точки в формате GeoJSON.
func (p *Point) GeoPoint() interface{} {
	if p == nil {
		return nil
	}
	return &struct {
		Type        string
		Coordinates []float64
	}{
		Type:        "Point",
		Coordinates: p[:],
	}
}

// GeoPolygon возвращает представление окружности с центорм в данной точке в виде полигона
// в формате GeoJSON.
//
// Все эти извращения нужны только для того, чтобы нормально работали поисковые индексы в MongoDB,
// по той простой причине, что Mongo позволяет сохранять только объекты в формате GeoJSON, а те
// умники, которые разрабатывали данный стандарт, решили, что круг — это абсолютно лишняя
// сущность, которая не вписывается в их мироощущение.
func (p *Point) GeoPolygon(radius float64) interface{} {
	if p == nil || radius <= 0 {
		return nil
	}
	const count = 12
	coords := make([][]float64, count+1)
	for i := range coords {
		coords[i] = p.Move(radius, 360.0/float64(count)*float64(i))[:]
	}
	return &struct {
		Type        string
		Coordinates [][]float64
	}{
		Type:        "Polygon",
		Coordinates: coords,
	}
}

const (
	EarthRadius float64 = 6378137.0 // радиус Земли в метрах
)

// Move возвращает новую точку, перемещенную от изначально на dist метрах в направлении bearing
// в градусах.
func (p *Point) Move(dist float64, bearing float64) *Point {
	dr := dist / EarthRadius
	bearing = bearing * math.Pi / 180.0
	lon1 := p.Longitude() * math.Pi / 180.0
	lat1 := p.Latitude() * math.Pi / 180.0
	lat2_part1 := math.Sin(lat1) * math.Cos(dr)
	lat2_part2 := math.Cos(lat1) * math.Sin(dr) * math.Cos(bearing)
	lat2 := math.Asin(lat2_part1 + lat2_part2)
	lon2_part1 := math.Sin(bearing) * math.Sin(dr) * math.Cos(lat1)
	lon2_part2 := math.Cos(dr) - (math.Sin(lat1) * math.Sin(lat2))
	lon2 := lon1 + math.Atan2(lon2_part1, lon2_part2)
	lon2 = math.Mod((lon2+3*math.Pi), (2*math.Pi)) - math.Pi
	lon2 = lon2 * 180.0 / math.Pi
	lat2 = lat2 * 180.0 / math.Pi
	return NewPoint(lon2, lat2)
}

// BearingTo возвращает направление в градусах на указанную точку.
func (p *Point) BearingTo(p2 *Point) float64 {
	dLon := (p2.Longitude() - p.Longitude()) * math.Pi / 180.0
	lat1 := p.Latitude() * math.Pi / 180.0
	lat2 := p2.Latitude() * math.Pi / 180.0
	y := math.Sin(dLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLon)
	brng := math.Atan2(y, x) * 180.0 / math.Pi
	return brng
}

// Distance возвращает дистанцию между двумя точками в метрах.
func (p *Point) Distance(p2 *Point) float64 {
	dLon := (p2.Longitude() - p.Longitude()) * math.Pi / 180.0
	dLat := (p2.Latitude() - p.Latitude()) * math.Pi / 180.0
	lat1 := p.Latitude() * math.Pi / 180.0
	lat2 := p2.Latitude() * math.Pi / 180.0
	a1 := math.Sin(dLat/2) * math.Sin(dLat/2)
	a2 := math.Sin(dLon/2) * math.Sin(dLon/2) * math.Cos(lat1) * math.Cos(lat2)
	a := a1 + a2
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
}
