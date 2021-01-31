package parse

import (
	"strconv"
	"testing"
)

func TestFormatFatalMarkerOneTemplateLine(t *testing.T) {
	template := []string{"line 1"}

	message := "message"
	sourceLineIndex := 0
	fatalMarker := &fatalMarker{message: message, fatalLine: sourceMarker{lineContents: "", sourceLineIndex: sourceLineIndex}}

	result := formatFatalMarker(fatalMarker, template)

	expected := "Fatal error " + message + " on line " + strconv.Itoa(sourceLineIndex+1) + ":\n"
	expected = expected + "1 > line 1"

	if result != expected {
		t.Error("Expected fatal message:\n|" + expected + "|\ngot\n|" + result + "|")
	}
}

func TestFormatFatalMarkerTwoTemplateLinesErrorOnFirst(t *testing.T) {
	template := []string{"line 1", "line 2"}

	message := "message"
	sourceLineIndex := 0
	fatalMarker := &fatalMarker{message: message, fatalLine: sourceMarker{lineContents: "", sourceLineIndex: sourceLineIndex}}

	result := formatFatalMarker(fatalMarker, template)

	expected := "Fatal error " + message + " on line " + strconv.Itoa(sourceLineIndex+1) + ":\n"
	expected = expected + "1 > line 1" + "\n"
	expected = expected + "2   line 2"

	if result != expected {
		t.Error("Expected fatal message:\n|" + expected + "|\ngot\n|" + result + "|")
	}
}

func TestFormatFatalMarkerTwoTemplateLineErrorOnSecond(t *testing.T) {
	template := []string{"line 1", "line 2"}

	message := "message"
	sourceLineIndex := 1
	fatalMarker := &fatalMarker{message: message, fatalLine: sourceMarker{lineContents: "", sourceLineIndex: sourceLineIndex}}

	result := formatFatalMarker(fatalMarker, template)

	expected := "Fatal error " + message + " on line " + strconv.Itoa(sourceLineIndex+1) + ":\n"
	expected = expected + "1   line 1" + "\n"
	expected = expected + "2 > line 2"

	if result != expected {
		t.Error("Expected fatal message:\n|" + expected + "|\ngot\n|" + result + "|")
	}
}

func TestFormatFatalMarkerThreeTemplateLinesErrorOnSecond(t *testing.T) {
	template := []string{"line 1", "line 2", "line 3"}

	message := "message"
	sourceLineIndex := 1
	fatalMarker := &fatalMarker{message: message, fatalLine: sourceMarker{lineContents: "", sourceLineIndex: sourceLineIndex}}

	result := formatFatalMarker(fatalMarker, template)

	expected := "Fatal error " + message + " on line " + strconv.Itoa(sourceLineIndex+1) + ":\n"
	expected = expected + "1   line 1" + "\n"
	expected = expected + "2 > line 2" + "\n"
	expected = expected + "3   line 3"

	if result != expected {
		t.Error("Expected fatal message:\n|" + expected + "|\ngot\n|" + result + "|")
	}
}

func TestFormatFatalMarkerOnEmptyLine(t *testing.T) {
	message := "Fatal error no line"
	fatalMarker := &fatalMarker{message: message, fatalLine: emptyLine}

	template := []string{""}
	result := formatFatalMarker(fatalMarker, template)

	expected := "Fatal error " + message
	if result != expected {
		t.Error("Expected fatal message:\n|" + expected + "|\ngot\n|" + result + "|")
	}
}
