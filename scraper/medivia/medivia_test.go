package medivia

import (
	"fmt"
	"testing"
)

func TestMedivia_WhoIs(t *testing.T) {
	m := New()
	c, err := m.WhoIs("Argaiv")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(c)
}
