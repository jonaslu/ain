package disk

import (
	"io"
	"os"
	"os/exec"

	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

const fallbackEditor = "vim"

func captureEditorOutput(tempFile *os.File) (string, error) {
	editorEnvVarName := "VISUAL"
	editorEnvStr := os.Getenv(editorEnvVarName)

	if editorEnvStr == "" {
		editorEnvVarName = "EDITOR"
		editorEnvStr = os.Getenv(editorEnvVarName)
	}

	if editorEnvStr == "" {
		_, err := exec.LookPath(fallbackEditor)
		if err != nil {
			return "", errors.New("cannot find the fallback editor vim on the $PATH. Cannot edit file.")
		}

		editorEnvVarName = fallbackEditor
		editorEnvStr = fallbackEditor
	}

	editorCmdAndArgs, err := utils.TokenizeLine(editorEnvStr)
	if err != nil {
		return "", errors.Wrapf(err, "cannot parse $%s environment variable", editorEnvVarName)
	}

	editorArgs := append(editorCmdAndArgs[1:], tempFile.Name())

	cmd := exec.Command(editorCmdAndArgs[0], editorArgs...)
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", errors.Wrap(err, "can't open /dev/tty")
	}

	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty

	err = cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, "error running $%s %s", editorEnvVarName, cmd.String())
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return "", errors.Wrap(err, "cannot seek template temp-file to 0")
	}

	tempFileContents, err := io.ReadAll(tempFile)
	if err != nil {
		return "", errors.Wrap(err, "cannot read from template temp-file")
	}

	return string(tempFileContents), nil
}

func readEditedRawTemplateString(sourceTemplateFileName string) (string, error) {
	rawTemplateString, err := os.Open(sourceTemplateFileName)
	if err != nil {
		return "", errors.Wrapf(err, "cannot open source template file %s", sourceTemplateFileName)
	}

	// .ini formats it like ini file in some editors
	tempFile, err := os.CreateTemp("", "ain*.ini")
	if err != nil {
		return "", errors.Wrap(err, "cannot create template temp-file")
	}

	defer func() {
		if removeErr := os.Remove(tempFile.Name()); removeErr != nil {
			wrappedRemoveErr := errors.Wrapf(removeErr, "could not remove template temp-file %s\nPlease delete it manually.", tempFile.Name())

			if err != nil {
				err = utils.CascadeErrorMessage(err, wrappedRemoveErr)
			} else {
				err = wrappedRemoveErr
			}
		}
	}()

	_, err = io.Copy(tempFile, rawTemplateString)
	if err != nil {
		return "", errors.Wrap(err, "cannot copy source template file to temp-file")
	}

	return captureEditorOutput(tempFile)
}

func ReadRawTemplateString(templateFileName string, editFile bool) (string, error) {
	if editFile {
		return readEditedRawTemplateString(templateFileName)
	}

	fileContents, err := os.ReadFile(templateFileName)
	if err != nil {
		return "", errors.Wrapf(err, "could not read template file %s", templateFileName)
	}

	return string(fileContents), nil

}
