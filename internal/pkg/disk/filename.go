package disk

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/jonaslu/ain/internal/pkg/utils"
	"github.com/pkg/errors"
)

func GetTemplateFilenames() ([]string, error) {
	var localTemplateFilenames []string

	if len(flag.Args()) >= 1 {
		localTemplateFilenames = flag.Args()
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "Could not stat stdin")
	}

	if (fi.Mode() & os.ModeCharDevice) == 0 {
		fileNameBytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.Wrap(err, "Could not read pipe stdin")
		}

		localTemplateFilenamesViaPipe, err := utils.TokenizeLine(string(fileNameBytes))
		if err != nil {
			return nil, errors.Wrap(err, "Could not read template name(s) from pipe")
		}

		if len(localTemplateFilenamesViaPipe) == 0 {
			return nil, errors.New("Pipe input did not contain any template names")
		}

		localTemplateFilenames = append(localTemplateFilenames, localTemplateFilenamesViaPipe...)
	}

	return localTemplateFilenames, nil
}
