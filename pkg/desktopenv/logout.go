package desktopenv

import (
	"errors"

	"github.com/godbus/dbus/v5"
)

// Logout logs the user out of the desktop environment.
// Supports GNOME and KDE.
func (d *DesktopEnv) Logout() error {
	// Try GNOME SessionManager
	obj := d.conn.Object("org.gnome.SessionManager", "/org/gnome/SessionManager")
	// Logout(mode uint32): 0 = normal, 1 = no confirmation, 2 = force
	call := obj.Call("org.gnome.SessionManager.Logout", 0, uint32(2))
	if call.Err == nil {
		return nil
	}

	// Try KDE KSMServer
	obj = d.conn.Object("org.kde.ksmserver", "/KSMServer")
	// logout(confirmMode int32, shutdownType int32, shutdownMode int32)
	// confirmMode: 0 = prompt, 1 = no confirm, 2 = force
	// shutdownType: 0 = logout, 1 = shutdown, 2 = reboot
	// shutdownMode: 0 = wait, 1 = try now, 2 = force now
	call = obj.Call("org.kde.KSMServerInterface.logout", 0, int32(2), int32(0), int32(0))
	if call.Err == nil {
		return nil
	}

	// Try freedesktop login1 as fallback
	sysConn, err := dbus.ConnectSystemBus()
	if err == nil {
		defer sysConn.Close()
		obj = sysConn.Object("org.freedesktop.login1", "/org/freedesktop/login1")
		call = obj.Call("org.freedesktop.login1.Manager.TerminateSession", 0, "self")
		if call.Err == nil {
			return nil
		}
	}

	return errors.New("failed to logout: no supported desktop environment found")
}
