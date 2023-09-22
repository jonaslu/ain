package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/disk"
)

func Assemble(ctx context.Context, filenames []string) (*data.BackendInput, string, error) {
	allSectionedTemplates := []*SectionedTemplate{}
	fatals := []string{}

	for _, filename := range filenames {
		// !! TODO !! The file-name will be displayed as test.ain! <- Remove the exclamation-mark
		// when setting the file-name2.
		rawTemplateString, err := disk.ReadRawTemplateString(filename)
		if err != nil {
			return nil, "", err
		}

		if sectionedTemplate := NewSections(rawTemplateString, filename); sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		} else {
			allSectionedTemplates = append(allSectionedTemplates, sectionedTemplate)
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	for _, sectionedTemplate := range allSectionedTemplates {
		if sectionedTemplate.substituteEnvVars(); sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	totalConfig := data.Config{Timeout: data.TimeoutNotSet}

	for i := len(allSectionedTemplates) - 1; i >= 0; i-- {
		sectionedTemplate := allSectionedTemplates[i]
		config := sectionedTemplate.getConfig()

		if sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
			break
		}

		if totalConfig.Timeout == data.TimeoutNotSet {
			totalConfig.Timeout = config.Timeout
		}

		if totalConfig.QueryDelim == nil {
			totalConfig.QueryDelim = config.QueryDelim
		}

		if totalConfig.Timeout > data.TimeoutNotSet && totalConfig.QueryDelim != nil {
			break
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	spew.Dump(totalConfig)

	return nil, "", nil
}

func main() {
	flag.Parse()
	filenames := flag.Args()

	_, fatals, err := Assemble(context.TODO(), filenames)
	if err != nil {
		panic(err)
	}

	if fatals != "" {
		fmt.Println(fatals)
	}
}

// I want this: I want to traverse all templates for all sections.
// I can do that in-mem. Then I want to find the overwriting ones

// Transform env-vars - how do I do that?
// Find all sections for each section.

// We'll do this, go thru all, subst env shit.

// Then parse out the config and go thru all with executables (they can return empty shit?)
