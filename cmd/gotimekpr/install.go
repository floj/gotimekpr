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
		src, err := os.Open(os.Args[0])
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

	return &cli.Command{
		Name:    "install",
		Aliases: []string{"i"},
		Usage:   "installs the gotimekpr systemd user service and starts it",
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := copyBin(); err != nil {
				return err
			}

			if err := copySystemdService(); err != nil {
				return err
			}

			cmd := exec.CommandContext(ctx, "systemctl", "--user", "enable", "gotimekpr.service")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}

			cmd = exec.CommandContext(ctx, "systemctl", "--user", "start", "gotimekpr.service")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}
