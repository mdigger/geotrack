package geo

// Polygon описывает полигон.
type Polygon [][]*Point

// Geo возвращает описание полигона в формате GeoJSON
func (p *Polygon) Geo() interface{} {
	if p == nil {
		return nil
	}
	return &struct {
		Type        string
		Coordinates *Polygon
	}{
		Type:        "Polygon",
		Coordinates: p,
	}
}
