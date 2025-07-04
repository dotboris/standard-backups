package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/go-viper/mapstructure/v2"
)

const TIME_FORMAT = "2006-01-02_15-04-05Z07-00" // limited special chars

type Options struct {
	DestinationDir string `mapstructure:"destination-dir"`
}

var Backend = &proto.BackendImpl{
	Backup: func(req *proto.BackupRequest) error {
		var options Options
		err := mapstructure.Decode(req.RawOptions, &options)
		if err != nil {
			return err
		}

		dest := path.Join(options.DestinationDir, time.Now().Format(TIME_FORMAT))
		err = os.MkdirAll(dest, 0755)
		if err != nil {
			return err
		}
		err = rsync(req.Paths, dest)
		if err != nil {
			return err
		}
		return nil
	},
}

func main() {
	Backend.Execute()
}

func rsync(sources []string, dest string) error {
	args := []string{"-av"}
	args = append(args, sources...)
	args = append(args, dest)
	cmd := exec.Command("rsync", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	fmt.Printf("running rsync: %s\n", cmd.String())
	err := cmd.Run()
	return err
}
