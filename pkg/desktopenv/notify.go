package desktopenv

import (
	"github.com/godbus/dbus/v5"
)

const (
	notifyDbusInterface = "org.freedesktop.Notifications"
	notifyDbusPath      = "/org/freedesktop/Notifications"
)

// SendNotification sends a desktop notification with the given text.
// Works with both GNOME and KDE via the freedesktop.org notification spec.
func (d *DesktopEnv) SendNotification(text string) error {
	obj := d.conn.Object(notifyDbusInterface, notifyDbusPath)
	call := obj.Call(
		notifyDbusInterface+".Notify",
		0,
		"gokpr-bazite",       // app_name
		d.lastNotificationID, // replaces_id
		"",                   // app_icon (empty = default)
		"Screen Time Alert",  // summary
		text,                 // body
		[]string{},           // actions
		map[string]dbus.Variant{ // hints
			"urgency":   dbus.MakeVariant(byte(1)), // normal urgency
			"transient": dbus.MakeVariant(true),
		},
		int32(-1), // expire_timeout (-1 = default)
	)
	if call.Err != nil {
		return call.Err
	}

	if err := call.Store(&d.lastNotificationID); err != nil {
		return err
	}

	return nil
}
