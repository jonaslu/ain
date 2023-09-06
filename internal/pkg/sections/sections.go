package main

import (
	"flag"
	"regexp"
	"strings"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"github.com/jonaslu/ain/internal/pkg/disk"
)

type SourceMarker struct {
	LineContents    string
	SourceLineIndex int
}

type Sections struct {
	ConfigSection         []SourceMarker
	HostSection           []SourceMarker
	QuerySection          []SourceMarker
	HeadersSection        []SourceMarker
	MethodSection         []SourceMarker
	BodySection           []SourceMarker
	BackendSection        []SourceMarker
	BackendOptionsSection []SourceMarker
	DefaultVars           []SourceMarker

	rawTemplateLines []string
}

type capturedSection struct {
	heading                string
	headingSourceLineIndex int
	sectionLines           *[]SourceMarker
}

const knownSectionHeaders = "host|query|headers|method|body|config|backend|backendoptions|defaultvars"

var knownSectionsRe = regexp.MustCompile(`(?i)^\s*\[(` + knownSectionHeaders + `)\]\s*$`)
var unescapeKnownSectionsRe = regexp.MustCompile(`(?i)^\s*\\\[(` + knownSectionHeaders + `)\]\s*$`)

var removeTrailingCommendRegExp = regexp.MustCompile("#.*$")
var isCommentOrWhitespaceRegExp = regexp.MustCompile(`^\s*#|^\s*$`)

func getSectionHeading(rawTemplateLine string) string {
	matchedLine := knownSectionsRe.FindStringSubmatch(rawTemplateLine)

	if len(matchedLine) == 2 {
		return strings.ToLower(matchedLine[1])
	}

	return ""
}

func NewSections(rawTemplateString, filename string) *Sections {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")
	sections := Sections{rawTemplateLines: rawTemplateLines}

	capturedSections := []capturedSection{}
	currentSectionLines := &[]SourceMarker{}

	for sourceIndex, rawTemplateLine := range rawTemplateLines {
		if isCommentOrWhitespaceRegExp.MatchString(rawTemplateLine) {
			continue
		}

		trailingCommentsRemoved := removeTrailingCommendRegExp.ReplaceAllString(rawTemplateLine, "")

		if sectionHeading := getSectionHeading(trailingCommentsRemoved); sectionHeading != "" {
			currentSectionLines = &[]SourceMarker{}
			capturedSections = append(capturedSections, capturedSection{
				heading:                sectionHeading,
				headingSourceLineIndex: sourceIndex,
				sectionLines:           currentSectionLines,
			})

			continue
		}

		if unescapeKnownSectionsRe.MatchString(trailingCommentsRemoved) {
			trailingCommentsRemoved = strings.Replace(trailingCommentsRemoved, `\`, "", 1)
		}

		sourceMarker := SourceMarker{
			LineContents:    strings.TrimRightFunc(trailingCommentsRemoved, func(r rune) bool { return unicode.IsSpace(r) }),
			SourceLineIndex: sourceIndex,
		}

		*currentSectionLines = append(*currentSectionLines, sourceMarker)
	}

	spew.Dump(capturedSections)

	return &sections
}

func main() {
	flag.Parse()
	filenames := flag.Args()
	for _, filename := range filenames {
		template, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			panic(err)
		}

		NewSections(template, filename)
	}
}
