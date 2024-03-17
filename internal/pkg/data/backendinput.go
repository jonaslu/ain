package data

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

func (bi *BackendInput) CreateBodyTempFile() error {
	if len(bi.Body) == 0 {
		return nil
	}

	tempFileDir := ""

	if bi.PrintCommand {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "Could not get current working dir, cannot store any body temp-file")
		}

		tempFileDir = cwd
	}

	bodyStr := strings.Join(bi.Body, "\n")

	tmpFile, err := os.CreateTemp(tempFileDir, "ain-body")
	if err != nil {
		return errors.Wrap(err, "Could not create tempfile")
	}

	if _, err := tmpFile.Write([]byte(bodyStr)); err != nil {
		// This also returns an error, but the first is more significant
		// so ignore this, it's only a temp-file that will be deleted eventually
		_ = tmpFile.Close()

		return errors.Wrap(err, "Could not write to tempfile")
	}

	bi.TempFileName = tmpFile.Name()

	return nil
}

func (bi *BackendInput) RemoveBodyTempFile(forceDeletion bool) error {
	if bi.TempFileName == "" {
		return nil
	}

	if !forceDeletion && bi.LeaveTempFile {
		return nil
	}

	err := os.Remove(bi.TempFileName)
	bi.TempFileName = ""

	return errors.Wrap(err, "Could not remove file with [Body] contents")
}
