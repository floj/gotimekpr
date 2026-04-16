-- name: NewTrackingRecord :one
INSERT INTO
    tracking(created_at, updated_at)
VALUES
    (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING *;

-- name: UpdateTrackingDuration :one
UPDATE
    tracking
SET
    duration_sec = unixepoch('now', 'localtime') - unixepoch(created_at, 'localtime'),
    updated_at = CURRENT_TIMESTAMP
WHERE
    id = ? RETURNING *;

-- name: GetDurationForToday :one
SELECT
    COUNT(id) AS count,
    SUM(duration_sec) AS total
FROM
    tracking
WHERE
    DATE(created_at, 'localtime') = DATE('now', 'localtime');

-- name: GetWeekdayLimitToday :one
SELECT
    *
FROM
    weekday_limits
WHERE
    weekday = strftime('%w', 'now');

-- name: GetDateLimitToday :one
SELECT
    *
FROM
    date_limits
WHERE
    DATE(limit_date) = DATE('now');

-- name: AddToDateLimitToday :one
INSERT INTO
    date_limits(limit_date, limit_minutes)
VALUES
    (DATE('now', 'localtime'), ?) ON CONFLICT(limit_date) DO
UPDATE
SET
    limit_minutes = limit_minutes + excluded.limit_minutes,
    updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: SetDateLimitToday :one
INSERT INTO
    date_limits(limit_date, limit_minutes)
VALUES
    (DATE('now'), ?) ON CONFLICT(limit_date) DO
UPDATE
SET
    limit_minutes = excluded.limit_minutes,
    updated_at = CURRENT_TIMESTAMP RETURNING *;

-- name: RemoveDateLimitToday :exec
DELETE FROM
    date_limits
WHERE
    DATE(limit_date, 'localtime') = DATE('now', 'localtime');