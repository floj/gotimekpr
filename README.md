# gotimekpr

A screen time tracking and limiting daemon for Linux desktops, inspired by [timekpr-next](https://launchpad.net/timekpr-next). It monitors active screen usage and enforces configurable daily time limits per weekday, logging out users when their allocated time is exceeded.

## Features

- Per-weekday screen time limits
- Automatic screen lock detection (pauses tracking while locked)
- Desktop notifications when approaching the time limit
- Automatic logout when the limit is exceeded
- Supports GNOME and KDE desktop environments
- Runs as a systemd user service

## Requirements

- Linux with a GNOME or KDE desktop environment
- Go 1.26+ (to build from source)
- D-Bus session bus

## Build

```sh
go build -o gotimekpr ./cmd/gotimekpr
```

No CGO is required — the project uses a pure Go SQLite driver.

## Install

Copy the binary and enable the systemd user service:

```sh
cp gotimekpr ~/.local/bin/
cp gotimekpr.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now gotimekpr
```

## Usage

```sh
# Run the daemon
gotimekpr daemon

# Run without automatic logout (for testing)
gotimekpr daemon --no-logout
```

## Configuration

The database and configuration are stored in `~/.config/gotimekpr/`. The directory is created automatically on first run.

Screen time limits are configured per weekday in the `limits` table of the SQLite database (`~/.config/gotimekpr/gotimekpr.db`). Set `duration_ms` to the allowed milliseconds per day, or `-1` for no limit.

| Weekday | Value |
|---------|-------|
| Sunday  | 0     |
| Monday  | 1     |
| Tuesday | 2     |
| ...     | ...   |
| Saturday| 6     |

## Development

Install [modd](https://github.com/cortesi/modd) for live-reload during development:

```sh
modd
```

This watches for file changes, regenerates SQL code with `sqlc`, and restarts the daemon with `--no-logout`.

## How It Works

1. The daemon polls every 5 seconds
2. Checks if the screen is locked via D-Bus — skips tracking if locked
3. Records active screen time in a local SQLite database
4. Compares daily totals against the configured limit for the current weekday
5. Sends a notification when less than 60 seconds remain
6. Logs out the user when the limit is exceeded

## License

[MIT](LICENSE)
