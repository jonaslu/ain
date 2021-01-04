package sections

import (
	"strings"
	"testing"

	"github.com/jonaslu/ain/template"
)

func TestCaptureSeveralOfSameSection(t *testing.T) {
	sectionName := "Host"

	templateStr := `
[` + sectionName + `]
[` + sectionName + `]`

	templat := template.TokenizeTemplate(templateStr)

	_, err := captureSection(sectionName, templat, false)

	if err == nil {
		t.Errorf("Expected error")
	}

	if !strings.Contains(err.Message, "Several ["+sectionName+"] sections found") {
		t.Error("Expected warning on several [" + sectionName + "]")
	}
}

func TestCaptureWithTrim(t *testing.T) {
	sectionName := "Host"

	templateStr := `
[` + sectionName + `]
   trim me`

	templat := template.TokenizeTemplate(templateStr)

	captureResult, err := captureSection(sectionName, templat, true)

	if err != nil {
		t.Errorf("Expected no errors")
	}

	if len(captureResult.sectionLines) != 1 {
		t.Errorf("Expected only one captured line")
	}

	capturedTrimmedLine := captureResult.sectionLines[0]

	if capturedTrimmedLine.LineContents != "trim me" {
		t.Errorf("Expected line to be trimmed: |%s|", capturedTrimmedLine.LineContents)
	}
}
