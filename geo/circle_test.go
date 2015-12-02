package geo

import (
	"testing"

	"github.com/kr/pretty"
)

func TestCircle(t *testing.T) {
	circle := &Circle{
		Center: NewPoint(37.57351, 55.715084),
		Radius: 500,
	}
	pretty.Println(circle.Geo())
}
