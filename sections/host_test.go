package sections

import (
	"strings"
	"testing"

	"github.com/jonaslu/ain/template"
)

func TestParseHostTwoHeaders(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
  [Host]`)

	warnings, error := ParseHostSection(templat, parsedTemplate)

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings")
	}

	if error == nil {
		t.Error("Expected one error on parsing [Host]")
	}

	if !strings.Contains(error.Message, "Several [Host] sections found") {
		t.Error("Expected warning on several [Hosts]")
	}
}

func TestParseHostEmptyHeader(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]`)

	warnings, error := ParseHostSection(templat, parsedTemplate)

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings")
	}

	if error == nil {
		t.Error("Expected one error on parsing [Host]")
	}

	if !strings.Contains(error.Message, "Empty [Host] line") {
		t.Error("Expected error on empty [Host]")
	}
}

func TestParseHostEmptyTwoHeaders(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
[Goat]`)

	warnings, error := ParseHostSection(templat, parsedTemplate)

	if len(warnings) != 0 {
		t.Errorf("Expected no warnings")
	}

	if error == nil {
		t.Error("Expected one error on parsing [Host]")
	}

	if !strings.Contains(error.Message, "Empty [Host] line") {
		t.Error("Expected error on empty [Host]")
	}
}

func TestParseHostTwoHostLines(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
http://localhost:8080/
http://localhost:8081/


[Goat]`)

	warnings, error := ParseHostSection(templat, parsedTemplate)

	if error != nil {
		t.Error("Expected no errors")
	}

	if len(warnings) != 2 {
		t.Errorf("Expected two warnings")
	}

	if !strings.Contains(warnings[0].Message, "Found several lines under [Host]") {
		t.Error("Did not get warning on multiple [Host] sections")
	}
}

func TestParseHostMalformedUrl(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
://cheeze


[Goat]`)

	warnings, error := ParseHostSection(templat, parsedTemplate)

	if len(warnings) != 0 {
		t.Error("Expected no warnings")
	}

	if error == nil {
		t.Error("Expected one error")
	}

	if !strings.Contains(error.Message, "Could not parse [Host] url") {
		t.Error("Parsing of url was correct")
	}
}

func TestParseHostHappyPath(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
http://localhost:8080/


[Goat]`)

	warnings, error := ParseHostSection(templat, parsedTemplate)

	if error != nil || len(warnings) != 0 {
		t.Error("Expected no errors or warnings")
	}

	if parsedTemplate.Host.Host != "localhost:8080" {
		t.Error("Host not parsed correctly")
	}
}
