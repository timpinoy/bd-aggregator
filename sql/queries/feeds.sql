-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
           $1,
           $2,
           $3,
           $4,
            $5,
            $6
       )
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedByID :one
SELECT * FROM feeds WHERE id = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
   SET updated_at = CURRENT_TIMESTAMP,
       last_fetched_at = CURRENT_TIMESTAMP
 WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT *
  FROM feeds
 ORDER BY last_fetched_at DESC NULLS FIRST
  LIMIT 1;