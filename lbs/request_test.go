package lbs

import (
	"testing"

	"github.com/kr/pretty"
)

var strs = []string{
	"864078-35827-010003698-fa-2-1e50-772a-95-1e50-773c-a6-1e50-7728-a1-1e50-7725-92-1e50-772d-90-1e50-7741-90-1e50-7726-88",
	"3D867293-14627-022970357-fa-2-1e3f-57f5-99-1e3f-9b10-86-1e3f-74ce-7f-1e3f-b6a6-7c-1e3f-6aaa-79-1e3f-57f6-78-1e3f-5100-77",
}

func TestParse(t *testing.T) {
	for _, s := range strs {
		data, err := Parse(s)
		if err != nil {
			t.Fatal(err)
		}
		pretty.Println(data)
	}
}
