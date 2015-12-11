package geo

// Polygon описывает полигон.
type Polygon [][]Point

// NewPolygon возвращает новое описание полигона для заданных точек.
func NewPolygon(points ...Point) Polygon {
	polygon := make([]Point, len(points))
	for i, point := range points {
		polygon[i] = point
	}
	p1, p2 := polygon[0], polygon[len(polygon)-1]
	if p1.Longitude() != p2.Longitude() || p1.Latitude() != p2.Latitude() {
		polygon = append(polygon, p1)
	}
	return Polygon{polygon}
}

// Geo возвращает описание полигона в формате GeoJSON
func (p Polygon) Geo() interface{} {
	if len(p) == 0 {
		return nil
	}
	return struct {
		Type        string
		Coordinates Polygon
	}{
		Type:        "Polygon",
		Coordinates: p,
	}
}
