package desktopenv

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strconv"

	"github.com/godbus/dbus/v5"
)

type gnomeDesktopEnv struct {
	conn               *dbus.Conn
	lastNotificationID uint32
}

// Logout logs the user out of the desktop environment.
// Supports GNOME and KDE, falls back to loginctl.
func (d *gnomeDesktopEnv) Logout() error {
	{
		// Try GNOME SessionManager
		obj := d.conn.Object("org.gnome.SessionManager", "/org/gnome/SessionManager")
		// Logout(mode uint32): 0 = normal, 1 = no confirmation, 2 = force
		call := obj.Call("org.gnome.SessionManager.Logout", 0, uint32(2))
		if call.Err == nil {
			return nil
		}
		slog.Warn("GNOME logout failed", "error", call.Err)
	}

	{
		// Try freedesktop login1 as fallback
		sysConn, err := dbus.ConnectSystemBus()
		if err == nil {
			defer sysConn.Close()
			obj := sysConn.Object("org.freedesktop.login1", "/org/freedesktop/login1")
			call := obj.Call("org.freedesktop.login1.Manager.TerminateSession", 0, "self")
			if call.Err == nil {
				return nil
			}
			slog.Warn("freedesktop logout failed", "error", call.Err)
		} else {
			slog.Warn("failed to connect to system bus for logout", "error", err)
		}
	}

	{
		// If all attempts fail, try loginctl as a last resort
		c := exec.Command("loginctl", "kill-user", strconv.Itoa(os.Getuid()))
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		err := c.Run()
		if err == nil {
			return nil
		}
		slog.Warn("loginctl logout failed", "error", err)
	}

	return errors.New("failed to logout: no supported desktop environment found")
}

func (d *gnomeDesktopEnv) SendNotification(title, message string) error {
	nid, err := sendNotification(d.conn, title, message, d.lastNotificationID)
	if err != nil {
		return err
	}
	d.lastNotificationID = nid
	return nil
}

func (d *gnomeDesktopEnv) IsScreenLocked() bool {
	// GNOME ScreenSaver
	if locked, ok := checkScreenSaver(d.conn, "org.gnome.ScreenSaver", "/org/gnome/ScreenSaver"); ok {
		return locked
	}

	// Freedesktop interface
	if locked, ok := checkScreenSaver(d.conn, "org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver"); ok {
		return locked
	}

	return false
}
