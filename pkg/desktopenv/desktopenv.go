package desktopenv

import (
	"github.com/godbus/dbus/v5"
)

type DesktopEnv struct {
	conn               *dbus.Conn
	lastNotificationID uint32
}

func New(conn *dbus.Conn) (*DesktopEnv, error) {
	return &DesktopEnv{conn: conn}, nil
}
