package desktopenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/godbus/dbus/v5"
)

type DesktopEnv interface {
	Logout() error
	SendNotification(title, message string) error
	IsScreenLocked() bool
}

func New(conn *dbus.Conn) (DesktopEnv, error) {
	flvr := flavor()
	switch flvr {
	case GNOME:
		return &gnomeDesktopEnv{}, nil
	case KDE:
		return &kdeDesktopEnv{}, nil
	default:
		return nil, fmt.Errorf("unsupported desktop environment: %s", flvr)
	}
}

type deFlavor string

const (
	Unknown  deFlavor = "unknown"
	GNOME    deFlavor = "gnome"
	KDE      deFlavor = "kde"
	XFCE     deFlavor = "xfce"
	Cinnamon deFlavor = "cinnamon"
	MATE     deFlavor = "mate"
	LXDE     deFlavor = "lxde"
	LXQT     deFlavor = "lxqt"
)

func flavor() deFlavor {
	desktop := ""
	for _, env := range []string{"XDG_CURRENT_DESKTOP", "XDG_SESSION_DESKTOP", "DESKTOP_SESSION"} {
		desktop = os.Getenv(env)
		if desktop != "" {
			desktop = strings.ToLower(desktop)
			break
		}
	}

	switch {
	case strings.Contains(desktop, "gnome"):
		return GNOME
	case strings.Contains(desktop, "kde"), strings.Contains(desktop, "plasma"):
		return KDE
	case strings.Contains(desktop, "xfce"):
		return XFCE
	case strings.Contains(desktop, "cinnamon"):
		return Cinnamon
	case strings.Contains(desktop, "mate"):
		return MATE
	case strings.Contains(desktop, "lxde"):
		return LXDE
	case strings.Contains(desktop, "lxqt"):
		return LXQT
	default:
		return Unknown
	}
}
