package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"
)

const TIME_FORMAT = "2006-01-02_15-04-05Z07-00" // limited special chars

func rsync(sources []string, dest string) error {
	args := []string{"-av"}
	args = append(args, sources...)
	args = append(args, dest)
	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func backup() error {
	paths, err := readPaths()
	if err != nil {
		return err
	}
	options, err := readOptions()
	if err != nil {
		return err
	}

	dest := path.Join(options.DestinationDir, time.Now().Format(TIME_FORMAT))
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}
	err = rsync(paths, dest)
	if err != nil {
		return err
	}

	return nil
}

func Execute() error {
	command, err := readCommand()
	if err != nil {
		return err
	}
	switch command {
	case "backup":
		return backup()
	default:
		return fmt.Errorf("unhandled command %s", command)
	}
}

func main() {
	err := Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
