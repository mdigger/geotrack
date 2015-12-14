package parser

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	for _, s := range testData {
		data, err := ParseLBS("gsm", s)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(data)
	}
}
