package sections

import (
	"testing"

	"github.com/jonaslu/ain/template"
)

func TestWarnDuplicateHeadersEntries(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Headers]
goat=cheese
  goat=cheese`)

	warnings, error := ParseHeadersSection(templat, parsedTemplate)

	if error != nil {
		t.Error("Expected no errors")
	}

	if len(warnings) != 1 {
		t.Error("Expected one warning")
	}

	if len(parsedTemplate.Headers) != 1 {
		t.Error("Expected one Headers entry")
	}

	if parsedTemplate.Headers[0] != "goat=cheese" {
		t.Error("Missing goat=cheese header entry")
	}
}

func TestParseHeadersTrimmed(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Headers]
    goat=cheese`)

	warnings, error := ParseHeadersSection(templat, parsedTemplate)

	if error != nil {
		t.Error("Expected no errors")
	}

	if len(warnings) != 0 {
		t.Error("Expected no warnings")
	}

	if len(parsedTemplate.Headers) != 1 {
		t.Error("Expected one Headers entry")
	}

	if parsedTemplate.Headers[0] != "goat=cheese" {
		t.Error("Entry goat=cheese not trimmed")
	}
}
