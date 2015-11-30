package lbs

import (
	"fmt"
	"testing"
)

var reqStr = "864078-35827-010003698-fa-2-1e50-772a-95-1e50-773c-a6-1e50-7728-a1-1e50-7725-92-1e50-772d-90-1e50-7741-90-1e50-7726-88"

func TestParse(t *testing.T) {
	data, err := Parse(reqStr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
