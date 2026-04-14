CREATE TABLE IF NOT EXISTS tracking (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tracking_created_at ON tracking (created_at);

CREATE TABLE IF NOT EXISTS limits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    weekday INTEGER NOT NULL UNIQUE,
    duration_ms INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- daily limits for each weekday, -1 means no limit, weekday 0 = Sunday, 1 = Monday, ..., 6 = Saturday
INSERT INTO
    limits (weekday, duration_ms)
VALUES
    -- Sunday  
    (0, -1),
    -- Monday
    (1, -1),
    -- Tuesday
    (2, 60 * 1000),
    -- Wednesday
    (3, -1),
    -- Thursday
    (4, -1),
    -- Friday
    (5, -1),
    -- Saturday
    (6, -1) ON conflict(weekday) DO nothing;