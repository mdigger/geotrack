package geo

import "math"

// CircleToPolygonSegments описывает количество сигментов, используемых для описания
// круга в виде полигона.
var CircleToPolygonSegments = 16

// Circle описывает круг с заданным радиусом.
type Circle struct {
	Center *Point
	Radius float64
}

// Geo возвращает описание круга в виде GeoJSON-объекта.
// По той простой идеи, что GeoJSON не поддерживает круги, он преобразуется в полигон.
// Количество элементов полигона задается глобальной переменной CircleToPolygonSegments.
func (c *Circle) Geo() interface{} {
	rLat := c.Radius / EarthRadius * 180.0 / math.Pi
	rLng := rLat / math.Cos(c.Center.Latitude()*math.Pi/180.0)
	dRad := 2.0 * math.Pi / float64(CircleToPolygonSegments)
	points := make([]*Point, CircleToPolygonSegments+1)
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
	return &struct {
		Type        string
		Coordinates Polygon
	}{
		Type:        "Polygon",
		Coordinates: Polygon{points},
	}
}
