package desktopenv

import (
	"github.com/godbus/dbus/v5"
)

// IsScreenLocked checks if the screen is locked on GNOME or KDE.
func (d *DesktopEnv) IsScreenLocked() bool {
	// Try GNOME ScreenSaver
	if locked, ok := checkScreenSaver(d.conn, "org.gnome.ScreenSaver", "/org/gnome/ScreenSaver"); ok {
		return locked
	}

	// Try KDE ScreenSaver (uses freedesktop interface)
	if locked, ok := checkScreenSaver(d.conn, "org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver"); ok {
		return locked
	}

	// Try KDE-specific interface
	if locked, ok := checkScreenSaver(d.conn, "org.kde.screensaver", "/ScreenSaver"); ok {
		return locked
	}

	return false
}

// checkScreenSaver queries a screensaver D-Bus service for lock status.
func checkScreenSaver(conn *dbus.Conn, dest, path string) (locked bool, ok bool) {
	obj := conn.Object(dest, dbus.ObjectPath(path))
	call := obj.Call(dest+".GetActive", 0)
	if call.Err != nil {
		return false, false
	}

	active := false
	if err := call.Store(&active); err != nil {
		return false, false
	}

	return active, true
}
