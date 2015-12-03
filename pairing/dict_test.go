package pairing

import (
	"fmt"
	"testing"
)

func TestDictionary(t *testing.T) {
	for i := 0; i < 100; i++ {
		fmt.Println(DictAlfa.Generate(4))
	}
}
