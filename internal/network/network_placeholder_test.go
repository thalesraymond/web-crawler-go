package network

import (
	"bytes"
	"testing"
)

func TestPlaceholder(t *testing.T) {
	var buf bytes.Buffer

	err := Placeholder(&buf)
	
	if err != nil {
		t.Error("Error writing to buffer: ", err)
	}

	expected := "Network Placeholder\n"
	
	if buf.String() != expected {
		t.Error("Buffer content does not match expected content: ", buf.String())
	}
}
