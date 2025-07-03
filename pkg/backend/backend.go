package backend

import (
	"fmt"
	"os"
)

type BackupFunc func(paths []string, options map[string]any) error

type Backend struct {
	Backup BackupFunc
}

func (b *Backend) execute() error {
	command, err := readCommand()
	if err != nil {
		return err
	}
	switch command {
	case "backup":
		if b.Backup == nil {
			return fmt.Errorf("unhandled command %s", command)
		}
		paths, err := readPaths()
		if err != nil {
			return err
		}
		options, err := readOptions()
		if err != nil {
			return err
		}
		return b.Backup(paths, options)
	default:
		return fmt.Errorf("unknown command %s", command)
	}
}

func (b *Backend) Execute() {
	err := b.execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
