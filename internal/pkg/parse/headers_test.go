package parse

import (
	"strings"
	"testing"

	"github.com/jonaslu/ain/internal/pkg/call"
)

func TestWarnDuplicateHeadersEntries(t *testing.T) {
	callData := &call.Data{}

	template, _ := trimTemplate(`
[Headers]
goat=cheese
  goat=cheese`)

	fatalMarker := parseHeadersSection(template, callData)

	if fatalMarker == nil {
		t.Error("Expected a fatalMarker")
	}

	if !strings.Contains(fatalMarker.message, "Same entry in [Headers] twice") {
		t.Error("Expected fatal message on same entry twice")
	}
}

func TestParseHeadersTrimmed(t *testing.T) {
	callData := &call.Data{}

	template, _ := trimTemplate(`
[Headers]
    goat=cheese`)

	fatalMarker := parseHeadersSection(template, callData)

	if fatalMarker != nil {
		t.Error("Expected no fatalMarker")
	}

	if len(callData.Headers) != 1 {
		t.Error("Expected one Headers entry")
	}

	if callData.Headers[0] != "goat=cheese" {
		t.Error("Entry goat=cheese not trimmed")
	}
}
