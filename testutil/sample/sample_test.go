package sample

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestGen(t *testing.T) {
	for i := 1; i < 122; i++ {
		addr, key := GenAccount()
		fmt.Printf("%d %s %s\n", i, addr.String(), hex.EncodeToString(key))
	}
}
