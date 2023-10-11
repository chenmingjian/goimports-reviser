package helper

import (
	"os"

	"github.com/chenmingjian/goimports-reviser/pkg/module"
	"github.com/chenmingjian/goimports-reviser/reviser"
)

func DetermineProjectName(projectName, filePath string) (string, error) {
	if filePath == reviser.StandardInput {
		var err error
		filePath, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	return module.DetermineProjectName(projectName, filePath)
}
