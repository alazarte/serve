package logger

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	var buffer bytes.Buffer

	Debugf("asd")
	b, err := io.ReadAll(&buffer)
	if err != nil {
		t.Fatalf("Failed to read buffer")
	}
	if strings.Contains(string(b), "asd") {
		t.Error("Shouldn't log debug")
	}

	buffer.Reset()

	Infof("asd")
	b, err = io.ReadAll(&buffer)
	if err != nil {
		t.Fatalf("Failed to read buffer")
	}
	if !strings.Contains(string(b), "asd") {
		t.Error("No substring")
	}
}
