package bosh

import (
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	"gopkg.in/yaml.v2"
)

type deleterFs interface {
	fileio.FileReader
	fileio.FileWriter
	fileio.TempDirer
}

type SSHKeyDeleter struct {
	stateStore stateStore
	fs         deleterFs
}

func NewSSHKeyDeleter(stateStore stateStore, fs deleterFs) SSHKeyDeleter {
	return SSHKeyDeleter{
		stateStore: stateStore,
		fs:         fs,
	}
}

func (s SSHKeyDeleter) Delete() error {
	varsDir, err := s.stateStore.GetVarsDir()
	if err != nil {
		return err
	}

	varsStore := filepath.Join(varsDir, "jumpbox-vars-store.yml")
	variables, err := s.fs.ReadFile(varsStore)
	if err == nil {
		varString, err := deleteJumpboxSSHKey(string(variables))
		if err != nil {
			return fmt.Errorf("Jumpbox variables: %s", err) //nolint:staticcheck
		}
		if string(variables) == varString {
			return nil
		}
		err = s.fs.WriteFile(varsStore, []byte(varString), storage.StateMode)
		if err != nil {
			//not tested
			return fmt.Errorf("Writing jumpbox vars store: %s", err) //nolint:staticcheck
		}
	}

	return nil
}

func deleteJumpboxSSHKey(varsString string) (string, error) {
	vars := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(varsString), &vars)
	if err != nil {
		return "", err
	}
	delete(vars, "jumpbox_ssh")
	newVars, err := yaml.Marshal(vars)
	if err != nil {
		return "", err // not tested
	}
	return string(newVars), nil
}
