package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

func printErrorAndExit(err error) {
	formattedError := fmt.Errorf("An error occurred: %v", err.Error())
	fmt.Fprintln(os.Stderr, formattedError.Error())
	os.Exit(1)
}

func captureEditorOutput(tempFile *os.File) string {
	editorCmd := os.Getenv("EDITOR")
	cmd := exec.Command(editorCmd, tempFile.Name())
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		errors.Wrap(err, "can't open /dev/tty")
	}

	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	err = cmd.Run()
	if err != nil {
		printErrorAndExit(err)
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		printErrorAndExit(err)
	}

	tempFileContents, err := ioutil.ReadAll(tempFile)
	if err != nil {
		printErrorAndExit(err)
	}

	return string(tempFileContents)
}

func copySourceTemplate(sourceTemplateFileName string) *os.File {
	sourceTemplate, err := os.Open(sourceTemplateFileName)
	if err != nil {
		printErrorAndExit(err)
	}

	// .ini formats it like ini file in some editors
	tempFile, err := ioutil.TempFile("", "ain*.ini")
	if err != nil {
		printErrorAndExit(err)
	}

	writtenLen, err := io.Copy(tempFile, sourceTemplate)
	if writtenLen == 0 {
		printErrorAndExit(errors.New("Written 0 bytes"))
	}

	return tempFile
}

func stripComments(editedTemplate string) string {
	strippedLines := []string{}

	allLines := strings.Split(editedTemplate, "\n")
	for _, line := range allLines {
		isCommentLine, _ := regexp.MatchString("^\\s*#", line)
		if !isCommentLine {
			strippedLines = append(strippedLines, line)
		}
	}

	return strings.Join(strippedLines, "\n")
}

func main() {
	fi, err := os.Stdin.Stat()
	if err != nil {
		printErrorAndExit(err)
	}

	var sourceTemplateFileName string
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		// Connected to a pipe
		fileNameBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			printErrorAndExit(err)
		}

		sourceTemplateFileName = string(fileNameBytes)
	} else {
		sourceTemplateFileName = os.Args[1]
	}

	tempFile := copySourceTemplate(sourceTemplateFileName)
	defer tempFile.Close()

	editedTemplate := captureEditorOutput(tempFile)

	fmt.Println("Tempfile contents", stripComments(editedTemplate))
}
