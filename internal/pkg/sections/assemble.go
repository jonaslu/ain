package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/call"
	"github.com/jonaslu/ain/internal/pkg/data"
	"github.com/jonaslu/ain/internal/pkg/disk"
	"github.com/jonaslu/ain/internal/pkg/utils"
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

	totalConfig := data.NewConfig()

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

	allExecutableAndArgs := []executableAndArgs{}
	for _, sectionedTemplate := range allSectionedTemplates {
		allExecutableAndArgs = append(allExecutableAndArgs, sectionedTemplate.captureExecutableAndArgs()...)

		if sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	allExecutablesOutput := callExecutables(ctx, totalConfig, allExecutableAndArgs)

	for _, sectionedTemplate := range allSectionedTemplates {
		if sectionedTemplate.insertExecutableOutput(&allExecutablesOutput); sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	var host, backend string

	for _, sectionedTemplate := range allSectionedTemplates {
		for _, hostSourceMarker := range *sectionedTemplate.GetNamedSection(HostSection) {
			host = host + hostSourceMarker.LineContents
		}

		backendSourceMarkers := *sectionedTemplate.GetNamedSection(BackendSection)
		if len(backendSourceMarkers) > 1 {
			sectionedTemplate.SetFatalMessage("Found several lines under [Backend]", backendSourceMarkers[0].SourceLineIndex)
		} else if len(backendSourceMarkers) == 1 {
			backendSourceMarker := backendSourceMarkers[0]
			requestedBackendName := strings.ToLower(backendSourceMarker.LineContents)

			if !call.ValidBackend(requestedBackendName) {
				foundMisspelledName := false
				for backendName, _ := range call.ValidBackends {
					if utils.LevenshteinDistance(requestedBackendName, backendName) < 3 {
						sectionedTemplate.SetFatalMessage(fmt.Sprintf("Unknown backend: %s. Did you mean %s", requestedBackendName, backendName), backendSourceMarker.SourceLineIndex)
						foundMisspelledName = true
					}
				}

				if !foundMisspelledName {
					sectionedTemplate.SetFatalMessage(fmt.Sprintf("Unknown backend %s", requestedBackendName), backendSourceMarker.SourceLineIndex)
				}
			}

			backend = requestedBackendName
		}

		if sectionedTemplate.HasFatalMessages() {
			fatals = append(fatals, sectionedTemplate.GetFatalMessages())
		}
	}

	if len(fatals) > 0 {
		return nil, strings.Join(fatals, "\n\n"), nil
	}

	if host == "" {
		fatals = append(fatals, "No mandatory [Host] section found")
	}

	if backend == "" {
		fatals = append(fatals, "No mandatory [Backend] section found")
	}

	if len(fatals) > 0 {
		// Since we no longer have a sectionedTemplate errors
		// are no longer linked to a file and we separate
		// with one newline
		return nil, strings.Join(fatals, "\n"), nil
	}

	fmt.Println(host, backend)

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
