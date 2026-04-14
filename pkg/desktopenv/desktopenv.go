package desktopenv

import (
	"github.com/godbus/dbus/v5"
)

type DesktopEnv struct {
	conn               *dbus.Conn
	lastNotificationID uint32
}

func New() (*DesktopEnv, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}

	return &DesktopEnv{conn: conn}, nil
}

func (d *DesktopEnv) Close() {
	d.conn.Close()
}
