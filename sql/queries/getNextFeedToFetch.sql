-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at DESC
LIMIT 1;