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

	parseResult := ParseHostSection(templat, parsedTemplate)

	if len(parseResult.warnings) != 0 {
		t.Errorf("Expected no warnings")
	}

	if len(parseResult.errors) != 1 {
		t.Error("Expected one error on parsing [Host]")
	}

	if !strings.Contains(parseResult.errors[0].Message, "Several [Host] sections found") {
		t.Error("Expected warning on several [Hosts]")
	}
}

func TestParseHostEmptyHeader(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]`)

	parseResult := ParseHostSection(templat, parsedTemplate)

	if len(parseResult.warnings) != 0 {
		t.Errorf("Expected no warnings")
	}

	if len(parseResult.errors) != 1 {
		t.Error("Expected one error on parsing [Host]")
	}

	if !strings.Contains(parseResult.errors[0].Message, "Empty [Host] line") {
		t.Error("Expected error on empty [Host]")
	}
}

func TestParseHostEmptyTwoHeaders(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
[Goat]`)

	parseResult := ParseHostSection(templat, parsedTemplate)

	if len(parseResult.warnings) != 0 {
		t.Errorf("Expected no warnings")
	}

	if len(parseResult.errors) != 1 {
		t.Error("Expected one error on parsing [Host]")
	}

	if !strings.Contains(parseResult.errors[0].Message, "Empty [Host] line") {
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

	parseResult := ParseHostSection(templat, parsedTemplate)

	if len(parseResult.errors) != 0 {
		t.Error("Expected no errors")
	}

	if len(parseResult.warnings) != 2 {
		t.Errorf("Expected two warnings")
	}

	if !strings.Contains(parseResult.warnings[0].Message, "Found several host lines") {
		t.Error("Found several host lines")
	}
}

func TestParseHostMalformedUrl(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
://cheeze


[Goat]`)

	parseResult := ParseHostSection(templat, parsedTemplate)

	if len(parseResult.warnings) != 0 {
		t.Error("Expected no warnings")
	}

	if len(parseResult.errors) != 1 {
		t.Error("Expected one error")
	}

	if !strings.Contains(parseResult.errors[0].Message, "Could not parse [Host] url") {
		t.Error("Parsing of url was correct")
	}
}

func TestParseHostHappyPath(t *testing.T) {
	parsedTemplate := &TemplateSections{}

	templat := template.TokenizeTemplate(`
[Host]
http://localhost:8080/


[Goat]`)

	parseResult := ParseHostSection(templat, parsedTemplate)

	if len(parseResult.errors) != 0 || len(parseResult.warnings) != 0 {
		t.Error("Expected no errors or warnings")
	}

	if parsedTemplate.Host.Host != "localhost:8080" {
		t.Error("Host not parsed correctly")
	}
}
