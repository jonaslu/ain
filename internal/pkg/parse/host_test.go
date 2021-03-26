package parse

import (
	"strings"
	"testing"

	"github.com/jonaslu/ain/internal/pkg/data"
)

func TestParseHostEmptyHeader(t *testing.T) {
	callData := &data.Data{}

	template, _ := trimTemplate(`
[Host]`)

	fatalMarker := parseHostSection(template, callData)

	if fatalMarker == nil {
		t.Error("Expected fatalMarker on parsing [Host]")
	}

	if !strings.Contains(fatalMarker.message, "Empty [Host] line") {
		t.Error("Expected fatal on empty [Host]")
	}
}

func TestParseHostEmptyTwoHeaders(t *testing.T) {
	callData := &data.Data{}

	template, _ := trimTemplate(`
[Host]
[Goat]`)

	fatalMarker := parseHostSection(template, callData)

	if fatalMarker == nil {
		t.Error("Expected fatalMarker on parsing [Host]")
	}

	if !strings.Contains(fatalMarker.message, "Empty [Host] line") {
		t.Error("Expected fatal message on empty [Host]")
	}
}

func TestParseHostMalformedUrl(t *testing.T) {
	callData := &data.Data{}

	template, _ := trimTemplate(`
[Host]
://cheeze


[Goat]`)

	fatalMarker := parseHostSection(template, callData)

	if fatalMarker == nil {
		t.Error("Expected one fatalMarker")
	}

	if !strings.Contains(fatalMarker.message, "Could not parse [Host] url") {
		t.Error("Parsing of url was correct")
	}
}

func TestParseHostHappyPath(t *testing.T) {
	callData := &data.Data{}

	template, _ := trimTemplate(`
[Host]
http://localhost:8080/


[Goat]`)

	fatalMarker := parseHostSection(template, callData)

	if fatalMarker != nil {
		t.Error("Expected no fatalMarker")
	}

	if callData.Host.Host != "localhost:8080" {
		t.Error("Host not parsed correctly")
	}
}
