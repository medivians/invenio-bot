package wiki

import (
	"fmt"
	"testing"
)

func TestWiki_WhereToSell(t *testing.T) {
	w := New()
	l := w.WhereToSell("demon shield")
	fmt.Print(l)
}
