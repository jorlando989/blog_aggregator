-- name: GetPostsForUser :many
SELECT *
FROM posts
WHERE feed_id IN (
    SELECT feed_id
    FROM feed_follows
    WHERE user_id = $1
)
ORDER BY published_at DESC
LIMIT $2;