package parse

import (
	"strings"
	"testing"
)

func TestCaptureSeveralOfSameSection(t *testing.T) {
	sectionName := "Host"

	templateStr := `
[` + sectionName + `]
[` + sectionName + `]`

	template, _ := trimTemplate(templateStr)

	_, fatalMarker := captureSection(sectionName, template, false)

	if fatalMarker == nil {
		t.Errorf("Expected a fatalMarker")
	}

	if !strings.Contains(fatalMarker.message, "Several ["+sectionName+"] sections found") {
		t.Error("Expected fatal on several [" + sectionName + "]")
	}
}

func TestCaptureWithTrim(t *testing.T) {
	sectionName := "Host"

	templateStr := `
[` + sectionName + `]
   trim me`

	template, _ := trimTemplate(templateStr)

	captureResult, fatalMarker := captureSection(sectionName, template, true)

	if fatalMarker != nil {
		t.Errorf("Expected no fatalMarkers")
	}

	if len(captureResult.sectionLines) != 1 {
		t.Errorf("Expected only one captured line")
	}

	capturedTrimmedLine := captureResult.sectionLines[0]

	if capturedTrimmedLine.lineContents != "trim me" {
		t.Errorf("Expected line to be trimmed: |%s|", capturedTrimmedLine.lineContents)
	}
}
