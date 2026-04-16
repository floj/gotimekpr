package desktopenv

import "github.com/godbus/dbus/v5"

const (
	notifyDbusInterface = "org.freedesktop.Notifications"
	notifyDbusPath      = "/org/freedesktop/Notifications"
)

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

func sendNotification(conn *dbus.Conn, title, message string, replaceID uint32) (uint32, error) {
	obj := conn.Object(notifyDbusInterface, notifyDbusPath)
	call := obj.Call(
		notifyDbusInterface+".Notify",
		0,
		"Screentime Limit", // app_name
		replaceID,          // replaces_id
		"",                 // app_icon (empty = default)
		title,              // summary
		message,            // body
		[]string{},         // actions
		map[string]dbus.Variant{ // hints
			"urgency":   dbus.MakeVariant(byte(1)), // normal urgency
			"transient": dbus.MakeVariant(true),
		},
		int32(-1), // expire_timeout (-1 = default)
	)
	if call.Err != nil {
		return 0, call.Err
	}

	notificationID := uint32(0)
	if err := call.Store(&notificationID); err != nil {
		return 0, err
	}

	return notificationID, nil
}
