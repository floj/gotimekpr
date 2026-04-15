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

-- name: GetWeekdayLimit :one
SELECT
    *
FROM
    weekday_limits
WHERE
    weekday = strftime('%w', 'now');

-- name: GetDateLimit :one
SELECT
    *
FROM
    date_limits
WHERE
    date = DATE('now');