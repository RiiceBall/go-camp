package service

import (
	"fmt"
	"testing"
)

func TestFormat(t *testing.T) {
	t.Log(fmt.Sprintf("%06d", 1))
}
