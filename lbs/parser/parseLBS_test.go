package parser

import (
	"testing"

	"github.com/kr/pretty"
)

func TestParser(t *testing.T) {
	for _, s := range testData {
		data, err := ParseLBS("gsm", s)
		if err != nil {
			t.Fatal(err)
		}
		pretty.Println(data)
	}
}
