package geo

import "math"

// CircleToPolygonSegments описывает количество сегментов, используемых для описания
// круга в виде полигона.
var CircleToPolygonSegments = 16

// Circle описывает круг с заданным радиусом.
type Circle struct {
	Center Point
	Radius float64
}

// NewCircle возвращает новое описание круга.
func NewCircle(lon, lat, radius float64) Circle {
	return Circle{
		Center: NewPoint(lon, lat),
		Radius: radius,
	}
}

// Polygon возвращает описание круга в виде полигона.
// Количество элементов полигона задается глобальной переменной CircleToPolygonSegments.
func (c Circle) Polygon() Polygon {
	rLat := c.Radius / EarthRadius * 180.0 / math.Pi
	rLng := rLat / math.Cos(c.Center.Latitude()*math.Pi/180.0)
	dRad := 2.0 * math.Pi / float64(CircleToPolygonSegments)
	points := make([]Point, CircleToPolygonSegments+1)
	for i := 0; i <= CircleToPolygonSegments; i++ {
		theta := dRad * float64(i)
		x := math.Cos(theta)
		if math.Abs(x) < 0.01 {
			x = 0.0
		}
		y := math.Sin(theta)
		if math.Abs(y) < 0.01 {
			y = 0.0
		}
		points[i] = NewPoint(c.Center.Longitude()+y*rLng, c.Center.Latitude()+x*rLat)
	}
	return NewPolygon(points...)
}

// Geo возвращает описание круга в виде GeoJSON-объекта.
// По той простой идеи, что GeoJSON не поддерживает круги, он преобразуется в полигон.
func (c Circle) Geo() interface{} {
	return struct {
		Type        string
		Coordinates Polygon
	}{
		Type:        "Polygon",
		Coordinates: c.Polygon(),
	}
}
