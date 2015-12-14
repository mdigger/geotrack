package geo

import (
	"fmt"
	"testing"
)

func TestCircle(t *testing.T) {
	circle := Circle{
		Center: NewPoint(37.57351, 55.715084),
		Radius: 500,
	}
	fmt.Println(circle.Geo())
}
