package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dotboris/standard-backups/pkg/proto"
	"github.com/go-viper/mapstructure/v2"
)

type Forget struct {
	Enable  bool
	Options map[string]any
}

type Options struct {
	Repo   string
	Forget Forget
	Env    map[string]string
}

var Backend = &proto.BackendImpl{
	Backup: func(req *proto.BackupRequest) error {
		var options Options
		err := mapstructure.Decode(req.RawOptions, &options)
		if err != nil {
			return err
		}

		exists, err := checkRepoExists(options.Repo, options.Env)
		if err != nil {
			return err
		}
		if !exists {
			fmt.Fprintf(os.Stderr, "repo %s does not exist, creating it", options.Repo)
			err := restic(options.Repo, options.Env, "init")
			if err != nil {
				return fmt.Errorf("failed to initialize repository %s: %w",
					options.Repo, err)
			}
		}

		tagArgs := []string{
			"--tag", fmt.Sprintf("sb:dest:%s", req.DestinationName),
			"--tag", fmt.Sprintf("sb:job:%s", req.JobName),
		}

		backupArgs := []string{"backup"}
		for _, exclude := range req.Exclude {
			backupArgs = append(backupArgs, "--exclude", exclude)
		}
		backupArgs = append(backupArgs, tagArgs...)
		backupArgs = append(backupArgs, req.Paths...)
		err = restic(options.Repo, options.Env, backupArgs...)
		if err != nil {
			return fmt.Errorf("failed to backup %v to repo %s: %w",
				req.Paths, options.Repo, err)
		}

		if options.Forget.Enable {
			forgetArgs := []string{"forget"}
			forgetArgs = append(forgetArgs, tagArgs...)
			forgetOptionArgs, err := optionsToArgs(options.Forget.Options)
			if err != nil {
				return err
			}
			forgetArgs = append(forgetArgs, forgetOptionArgs...)
			err = restic(options.Repo, options.Env, forgetArgs...)
			if err != nil {
				return fmt.Errorf("failed to forget %v to repo %s: %w",
					req.Paths, options.Repo, err)
			}
		}

		return nil
	},
	Exec: func(req *proto.ExecRequest) error {
		var options Options
		err := mapstructure.Decode(req.RawOptions, &options)
		if err != nil {
			return err
		}

		err = restic(options.Repo, options.Env, req.Args...)
		return err
	},
	ListBackups: func(req *proto.ListBackupsRequest) (*proto.ListBackupsResponse, error) {
		var options Options
		err := mapstructure.Decode(req.RawOptions, &options)
		if err != nil {
			return nil, err
		}

		bs, err := resticOutput(options.Repo, options.Env, "snapshots", "--json")
		if err != nil {
			return nil, err
		}

		type Summary struct {
			TotalBytesProcessed int `json:"total_bytes_processed"`
		}
		type Snapshot struct {
			RawId   string   `json:"raw_id"`
			Time    string   `json:"time"`
			Tags    []string `json:"tags"`
			Summary Summary  `json:"summary"`
		}

		var snapshots []Snapshot
		err = json.Unmarshal(bs, &snapshots)
		if err != nil {
			return nil, err
		}

		var rawSnapshots []map[string]any
		err = json.Unmarshal(bs, &rawSnapshots)
		if err != nil {
			return nil, err
		}

		backups := make([]proto.ListBackupsResponseItem, len(snapshots))
		for i, snap := range snapshots {
			job := ""
			dest := ""
			for _, tag := range snap.Tags {
				j, ok := strings.CutPrefix(tag, "sb:job:")
				if ok {
					job = j
				}
				d, ok := strings.CutPrefix(tag, "sb:dest:")
				if ok {
					dest = d
				}
			}

			backups[i] = proto.ListBackupsResponseItem{
				Id:          snap.RawId,
				Time:        snap.Time,
				Bytes:       snap.Summary.TotalBytesProcessed,
				Job:         job,
				Destination: dest,
				Extra:       rawSnapshots[i],
			}
		}

		return &proto.ListBackupsResponse{
			Backups: backups,
		}, nil
	},
}

func main() {
	Backend.Execute()
}
