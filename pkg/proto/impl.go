package proto

import (
	"fmt"
	"os"
)

type BackendImpl struct {
	Backup BackupFunc
	Exec   ExecFunc
}

func (bi *BackendImpl) execute() error {
	command, err := GetCommand()
	if err != nil {
		return err
	}
	switch command {
	case "backup":
		return bi.backup()
	case "exec":
		return bi.exec()
	default:
		return fmt.Errorf("unknown command %s", command)
	}
}

func (bi *BackendImpl) Execute() {
	err := bi.execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
