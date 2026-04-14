-- name: InsertTrackingRecord :one
INSERT INTO
    tracking(created_at, updated_at)
VALUES
    (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING *;

-- name: AddTrackingRecordDuration :one
UPDATE
    tracking
SET
    duration_ms = duration_ms + ?,
    updated_at = CURRENT_TIMESTAMP
WHERE
    id = ? RETURNING *;

-- name: GetDurationForToday :one
SELECT
    COUNT(id) AS count,
    SUM(duration_ms) AS total
FROM
    tracking
WHERE
    DATE(created_at) = DATE('now');

-- name: GetLimits :many
SELECT
    *
FROM
    limits;