package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	_ "embed"

	"github.com/urfave/cli/v3"
)

//go:embed gotimekpr.service
var systemdService string

func cmdInstall() *cli.Command {
	copyBin := func() error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		localBinDir := filepath.Join(homeDir, ".local", "bin")
		if err := os.MkdirAll(localBinDir, 0755); err != nil {
			return err
		}
		executablePath := filepath.Join(localBinDir, "gotimekpr")
		// copy the currently running executable to the local bin directory

		// check if we are running in the same path as the target path, if so, skip copying
		selfPath, err := os.Executable()
		if err != nil {
			return err
		}
		selfPath, err = filepath.EvalSymlinks(selfPath)
		if err != nil {
			return err
		}

		if selfPath == executablePath {
			return nil
		}

		src, err := os.Open(selfPath)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(executablePath)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		if err := os.Chmod(executablePath, 0755); err != nil {
			return err
		}
		return nil
	}

	copySystemdService := func() error {
		confDir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		systemdUserDir := filepath.Join(confDir, "systemd", "user")
		if err := os.MkdirAll(systemdUserDir, 0755); err != nil {
			return err
		}
		serviceFile := filepath.Join(systemdUserDir, "gotimekpr.service")
		if err := os.WriteFile(serviceFile, []byte(systemdService), 0644); err != nil {
			return err
		}
		return nil
	}

	runCmd := func(ctx context.Context, name string, args ...string) error {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "installs the gotimekpr systemd user service and starts it",
		Action: func(ctx context.Context, c *cli.Command) error {

			// stop the service if it's already running, ignore errors since it might not be installed yet
			_ = runCmd(ctx, "systemctl", "--user", "stop", "gotimekpr.service")

			if err := copyBin(); err != nil {
				return err
			}

			if err := copySystemdService(); err != nil {
				return err
			}

			if err := runCmd(ctx, "systemctl", "--user", "daemon-reload"); err != nil {
				return err
			}

			if err := runCmd(ctx, "systemctl", "--user", "enable", "--now", "gotimekpr.service"); err != nil {
				return err
			}

			return nil
		},
	}
}
