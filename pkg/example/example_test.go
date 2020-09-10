package example

import "testing"

func TestIt(t *testing.T) {
	if GetTwo() != 2 {
		t.Errorf("GetTwo() should equal 2")
	}
}
