package desktopenv

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strconv"

	"github.com/godbus/dbus/v5"
)

type kdeDesktopEnv struct {
	conn               *dbus.Conn
	lastNotificationID uint32
}

// Logout logs the user out of the desktop environment.
// Supports GNOME and KDE, falls back to loginctl.
func (d *kdeDesktopEnv) Logout() error {
	{
		// Try KDE Plasma 6 Shutdown interface
		obj := d.conn.Object("org.kde.ksmserver", "/Shutdown")
		call := obj.Call("org.kde.Shutdown.logout", 0)
		if call.Err == nil {
			return nil
		}
		slog.Warn("KDE Plasma 6 logout failed", "error", call.Err)

	}

	{
		// Try KDE Plasma 5 KSMServer (legacy)
		obj := d.conn.Object("org.kde.ksmserver", "/KSMServer")
		// logout(confirmMode int32, shutdownType int32, shutdownMode int32)
		// confirmMode: 0 = prompt, 1 = no confirm, 2 = force
		// shutdownType: 0 = logout, 1 = shutdown, 2 = reboot
		// shutdownMode: 0 = wait, 1 = try now, 2 = force now
		call := obj.Call("org.kde.KSMServerInterface.logout", 0, int32(2), int32(0), int32(0))
		if call.Err == nil {
			return nil
		}
		slog.Warn("KDE Plasma 5 logout failed", "error", call.Err)
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

func (d *kdeDesktopEnv) SendNotification(title, message string) error {
	nid, err := sendNotification(d.conn, title, message, d.lastNotificationID)
	if err != nil {
		return err
	}
	d.lastNotificationID = nid
	return nil
}

func (d *kdeDesktopEnv) IsScreenLocked() bool {
	//  KDE-specific interface
	if locked, ok := checkScreenSaver(d.conn, "org.kde.screensaver", "/ScreenSaver"); ok {
		return locked
	}

	// Freedesktop interface
	if locked, ok := checkScreenSaver(d.conn, "org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver"); ok {
		return locked
	}

	return false
}
