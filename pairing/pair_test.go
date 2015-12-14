package pairing

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestPairs(t *testing.T) {
	var pairs Pairs
	pairs.Dictionary = DictAlfa
	for i := 0; i < 20; i++ {
		deviceID := fmt.Sprintf("%02d", rand.Intn(50))
		key := pairs.Generate(deviceID)
		fmt.Println(deviceID, key, pairs.GetDeviceID(key))
	}
}
