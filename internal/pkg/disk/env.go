package disk

import (
	"os"

	"github.com/hashicorp/go-envparse"
	"github.com/pkg/errors"
)

func ReadEnvFile(path string, errorOnMissingFile bool) error {
	file, err := os.Open(path)

	if os.IsNotExist(err) {
		if errorOnMissingFile {
			return errors.New("cannot open .env-file " + path)
		}

		return nil
	}

	if err != nil {
		return errors.Wrap(err, "error loading .env-file "+path)
	}

	if file != nil {
		res, err := envparse.Parse(file)
		if err != nil {
			return errors.Wrap(err, "error parsing .env-file "+path)
		}

		for envVarKey, envVarValue := range res {
			if _, exists := os.LookupEnv(envVarKey); !exists {
				if err := os.Setenv(envVarKey, envVarValue); err != nil {
					return errors.Wrap(err, "error setting env value from .env-file "+path)
				}
			}
		}
	}

	return nil
}
