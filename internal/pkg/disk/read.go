package disk

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

const EDIT_FILE_SUFFIX = "!"

func captureEditorOutput(tempFile *os.File) (string, error) {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}

	cmd := exec.Command(editorCmd, tempFile.Name())
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", errors.Wrap(err, "Can't open /dev/tty")
	}

	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, "Error running $EDITOR %s", cmd.String())
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return "", errors.Wrap(err, "Cannot seek template temp-file to 0")
	}

	tempFileContents, err := ioutil.ReadAll(tempFile)
	if err != nil {
		return "", errors.Wrap(err, "Cannot read from template temp-file")
	}

	return string(tempFileContents), nil
}

func readEditedTemplate(sourceTemplateFileName string) (str string, err error) {
	sourceTemplate, err := os.Open(sourceTemplateFileName)
	if err != nil {
		return "", errors.Wrapf(err, "Cannot open source template file %s", sourceTemplateFileName)
	}

	// .ini formats it like ini file in some editors
	tempFile, err := ioutil.TempFile("", "ain*.ini")
	if err != nil {
		return "", errors.Wrap(err, "Cannot create template temp-file")
	}

	defer func() {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			wrappedRemoveErr := errors.Wrapf(removeErr, "Could not remove template temp-file %s\nPlease delete it manually.", tempFile.Name())

			if err != nil {
				err = utils.CascadeErrorMessage(err, wrappedRemoveErr)
			} else {
				err = wrappedRemoveErr
			}
		}
	}()

	writtenLen, err := io.Copy(tempFile, sourceTemplate)
	if writtenLen == 0 {
		return "", errors.Wrap(err, "Cannot copy source template file to temp-file")
	}

	return captureEditorOutput(tempFile)
}

func ReadTemplate(templateFileName string) (string, error) {
	if strings.HasSuffix(templateFileName, EDIT_FILE_SUFFIX) {
		return readEditedTemplate(strings.TrimSuffix(templateFileName, EDIT_FILE_SUFFIX))
	}

	fileContents, err := ioutil.ReadFile(templateFileName)
	if err != nil {
		return "", errors.Wrapf(err, "Could not read template file %s", templateFileName)
	}

	return string(fileContents), nil

}
