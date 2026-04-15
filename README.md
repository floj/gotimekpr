# gotimekpr

A screen time tracking and limiting daemon for Linux desktops, inspired by [timekpr-next](https://launchpad.net/timekpr-next). It monitors active screen usage and enforces configurable daily time limits per weekday, logging out users when their allocated time is exceeded.

## Motivation

I used [timekpr-next](https://launchpad.net/timekpr-next) to limit my 13-year-old's screen time. He's really into Minecraft and recently heard that Bazzite Linux is *the* best distro for gaming. Being the supportive dad I am, I encouraged him to try it out, and we installed it on his PC. Father of the year, right?

Plot twist: Bazzite is based on ostree, and there's no easy way to install timekpr-next. So in a classic case of "I'll just write my own," I spent some hours building a stripped-down version of timekpr-next in Go. It runs entirely out of the user's home directory and keeps all its state there. No root required and works nicely with ostree's read-only root fs. Peak parental engineering.

Now, if he figures out that he can:
- stop the systemd user service himself
- tweak the limits via the CLI or just go full hacker mode with `sqlite3`

...then honestly, it's fine and I'll be proud he figured it out. But until that day comes, the screen time throne is mine again.

## Features

- Per-weekday screen time limits with per-date overrides
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
./build.sh
```

Or manually:

```sh
CGO_ENABLED=0 go build -o gotimekpr ./cmd/gotimekpr
```

No CGO required, we use the modernc.org/sqlite driver.

## Install

Use the built-in install command:

```sh
./gotimekpr install
```

This copies the binary to `~/.local/bin/`, installs the systemd user service, and enables it.

## Usage

```sh
# Run the daemon
gotimekpr daemon

# Run without automatic logout (for testing)
gotimekpr daemon --no-logout

# Show today's usage
gotimekpr usage

# Show today's limit
gotimekpr limits get

# Set today's limit to 2 hours
gotimekpr limits set 2h

# Add 30 minutes to today's limit
gotimekpr limits add 30m

# Enable debug logging
gotimekpr --debug daemon
```

The `--no-logout` flag and `--debug` flag can also be set via the `GOTIMEKPR_NO_LOGOUT` and `GOTIMEKPR_DEBUG` environment variables.

## Configuration

The database is stored in `~/.local/state/gotimekpr/gotimekpr.db`. The directory is created automatically on first run.

Screen time limits are configured per weekday in the `weekday_limits` table. Set `limit_minutes` to the allowed minutes per day, or `-1` for no limit. All weekdays default to `-1` (unlimited).

Per-date overrides can be set in the `date_limits` table, which take priority over weekday limits. The `limits set` and `limits add` commands operate on today's date limit.

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

This watches for file changes, regenerates SQL code with [sqlc](https://sqlc.dev), and restarts the daemon with `--no-logout`.

## How It Works

1. The daemon polls every 3 seconds
2. Checks if the screen is locked via D-Bus and skips tracking if locked
3. Records active screen time in a local SQLite database
4. Compares daily totals against the configured limit for the current weekday (or date override)
5. Sends a notification when less than 60 seconds remain
6. Logs out the user when the limit is exceeded

## License

[MIT](LICENSE)
