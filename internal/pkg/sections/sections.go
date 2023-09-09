package main

import (
	"flag"
	"fmt"
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

	fatals []string

	rawTemplateLines []string
	filename         string
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

func checkValidHeadings(capturedSections []capturedSection, sections *Sections) {
	// Keeps "header": [1,5,7] <- Name of heading and on what lines in the file
	headingDefinitionSourceLines := map[string][]int{}

	for _, capturedSection := range capturedSections {
		if len(*capturedSection.sectionLines) == 0 {
			sections.SetFatalMessage(fmt.Sprintf("Empty %s section", capturedSection.heading), capturedSection.headingSourceLineIndex)
		}

		headingDefinitionSourceLines[capturedSection.heading] = append(headingDefinitionSourceLines[capturedSection.heading], capturedSection.headingSourceLineIndex)
	}

	// !! TODO !! We now know the sourceLineIndex where multiple headings
	// are defined so we can print this more nicely
	for heading, headingSourceLineIndex := range headingDefinitionSourceLines {
		if len(headingSourceLineIndex) > 1 {
			sections.fatals = append(
				sections.fatals,
				fmt.Sprintf("Several %s sections found on line %d and %d", heading, headingSourceLineIndex[0]+1, headingSourceLineIndex[1]+1))
		}
	}
}

func getCapturedSections(rawTemplateLines []string) ([]capturedSection, bool) {
	templateEmpty := true
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

			templateEmpty = false

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

	return capturedSections, templateEmpty
}

func NewSections(rawTemplateString, filename string) *Sections {
	rawTemplateLines := strings.Split(strings.ReplaceAll(rawTemplateString, "\r\n", "\n"), "\n")
	sections := Sections{
		rawTemplateLines: rawTemplateLines,
		filename:         filename,
	}

	capturedSections, templateEmpty := getCapturedSections(rawTemplateLines)

	if templateEmpty {
		// !! TODO !! Change to no valid headings found, it can be full of stuff
		sections.fatals = []string{"Cannot process empty template"}
	} else {
		checkValidHeadings(capturedSections, &sections)
	}

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

		sections := NewSections(template, filename)
		if sections.HasFatalMessages() {
			fmt.Println(sections.GetFatalMessages())
		} else {
			spew.Dump(sections)
		}
	}
}
