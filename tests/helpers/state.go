package helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func ClearLocalTFState(tfDir string) error {
	var err error
	tfDir, err = filepath.Abs(tfDir)

	if err != nil {
		return err
	}

	tfDir = filepath.Clean(tfDir)

	_, err = os.Stat(tfDir)

	if os.IsNotExist(err) {
		return err
	}

	if err := os.Remove(fmt.Sprintf("%s/.terraform/terraform.tfstate", tfDir)); err != nil {
		pathErr, ok := err.(*os.PathError)
		if !ok {
			return err
		}
		if errors.Is(pathErr.Err, syscall.ENOENT) {
			return nil
		}
		return pathErr
	}
	return nil
}
