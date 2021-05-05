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
			return errors.Errorf("cannot open .env-file on path: %s", path)
		}

		return nil
	}

	if err != nil {
		return errors.Wrap(err, "Error loading environment "+path+" file")
	}

	if file != nil {
		res, err := envparse.Parse(file)
		if err != nil {
			return errors.Wrap(err, "Error parsing "+path+" file")
		}

		for envVarKey, envVarValue := range res {
			if _, exists := os.LookupEnv(envVarKey); !exists {
				if err := os.Setenv(envVarKey, envVarValue); err != nil {
					return errors.Wrap(err, "Error setting env value from file "+path)
				}
			}
		}
	}

	return nil
}
