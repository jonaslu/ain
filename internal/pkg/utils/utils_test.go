package utils

import (
	"reflect"
	"strings"
	"testing"
)

func tokenizeAndExpectError(commandLine string, t *testing.T) {
	_, err := TokenizeLine(commandLine)
	if err == nil {
		t.Fatal("Expected error to be non-nil")
	}

	if !strings.Contains(err.Error(), "Unterminated quote sequence") {
		t.Fatal("Expected unterminated quote sequence error message")
	}
}

func tokenizeAndCompare(commandLine string, expected []string, t *testing.T) {
	res, err := TokenizeLine(commandLine)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("Result: %s, did not match expected: %s", strings.Join(res, "|"), strings.Join(expected, "|"))
	}
}

func TestEverythingPositive(t *testing.T) {
	commandLine := `""`
	tokenizeAndCompare(commandLine, []string{`""`}, t)

	commandLine = `" "`
	tokenizeAndCompare(commandLine, []string{`" "`}, t)

	commandLine = `"yaketi 😎 yak"`
	tokenizeAndCompare(commandLine, []string{`"yaketi 😎 yak"`}, t)

	commandLine = `''`
	tokenizeAndCompare(commandLine, []string{`''`}, t)

	commandLine = `   'doh'`
	tokenizeAndCompare(commandLine, []string{`'doh'`}, t)

	commandLine = `goat`
	tokenizeAndCompare(commandLine, []string{`goat`}, t)

	commandLine = `   goat`
	tokenizeAndCompare(commandLine, []string{`goat`}, t)

	commandLine = `"this has \" unquote"   `
	tokenizeAndCompare(commandLine, []string{`"this has \" unquote"`}, t)

	commandLine = `  'and this \' "" has unquote'`
	tokenizeAndCompare(commandLine, []string{`'and this \' "" has unquote'`}, t)

	commandLine = `\"`
	tokenizeAndCompare(commandLine, []string{`\"`}, t)
}

func TestUnterminatedQuotes(t *testing.T) {
	tokenizeAndExpectError(`"`, t)
	tokenizeAndExpectError(`"\"`, t)
	tokenizeAndExpectError(`   "\"    `, t)

	tokenizeAndExpectError(`'`, t)
	tokenizeAndExpectError(`'\'`, t)
	tokenizeAndExpectError(`   '\'    `, t)
}
