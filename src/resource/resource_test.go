package resource

import (
	"fmt"
	"testing"
)

func TestNewRCB(t *testing.T) {
	rcb := NewRCB("R1", 1)
	fmt.Println(rcb)
}
