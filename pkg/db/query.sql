-- name: NewTrackingRecord :one
INSERT INTO
    tracking(created_at, updated_at)
VALUES
    (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING *;

-- name: UpdateTrackingDuration :one
UPDATE
    tracking
SET
    duration_sec = unixepoch('now') - unixepoch(created_at),
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
    DATE(created_at) = DATE('now');

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
    (DATE('now'), ?) ON CONFLICT(limit_date) DO
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
    DATE(limit_date) = DATE('now');