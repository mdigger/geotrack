package geo

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPoint(t *testing.T) {
	point := NewPoint(37.57351, 55.715084)
	data, err := json.Marshal(point)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(data))
	fmt.Println(point.Geo())
	var point2 Point
	err = json.Unmarshal(data, &point2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(point2)
	fmt.Println(point.Distance(NewPoint(37.589248, 55.765944)))

	p1 := NewPoint(-139.398, 77.1539)
	p2 := NewPoint(-139.55, -77.1804)
	fmt.Printf("%f\n", p1.Distance(p2))
}
