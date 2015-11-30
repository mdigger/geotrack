package geo

// GeoObject описывает гео-объект в формате GeoJSON
type GeoObject struct {
	Type        string
	Coordinates []float64
}
