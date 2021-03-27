package disk

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func captureEditorOutput(tempFile *os.File) (string, error) {
	// !! TODO !! If editorCmd is not set - warn and default to vim
	editorCmd := os.Getenv("EDITOR")

	cmd := exec.Command(editorCmd, tempFile.Name())
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", errors.Wrap(err, "can't open /dev/tty")
	}

	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, "error running command: %s", cmd.String())
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return "", errors.Wrap(err, "cannot seek tempfile to 0")
	}

	tempFileContents, err := ioutil.ReadAll(tempFile)
	if err != nil {
		return "", errors.Wrap(err, "cannot read from tempfile")
	}

	return string(tempFileContents), nil
}

func readEditedTemplate(sourceTemplateFileName string) (string, error) {
	sourceTemplate, err := os.Open(sourceTemplateFileName)
	if err != nil {
		return "", errors.Wrap(err, "cannot open source template file")
	}

	// .ini formats it like ini file in some editors
	tempFile, err := ioutil.TempFile("", "ain*.ini")
	if err != nil {
		return "", errors.Wrap(err, "cannot create tempfile")
	}
	defer tempFile.Close()

	writtenLen, err := io.Copy(tempFile, sourceTemplate)
	if writtenLen == 0 {
		return "", errors.Wrap(err, "cannot copy source file to temp-file")
	}

	return captureEditorOutput(tempFile)
}

func ReadTemplate(templateFileName string, execute bool) (string, error) {
	if execute {
		fileContents, err := ioutil.ReadFile(templateFileName)
		if err != nil {
			return "", errors.Wrapf(err, "Could not read file with name: %s", templateFileName)
		}

		return string(fileContents), nil
	}

	return readEditedTemplate(templateFileName)
}
