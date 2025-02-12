-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *,
    (SELECT name FROM users WHERE id = feed_follows.user_id) AS user_name,
    (SELECT name FROM feeds WHERE id = feed_follows.feed_id) AS feed_name;

-- name: RemoveFeedFollow :one
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1 AND feed_follows.feed_id = $2
RETURNING *,
    (SELECT name FROM users WHERE id = feed_follows.user_id) AS user_name,
    (SELECT name FROM feeds WHERE id = feed_follows.feed_id) AS feed_name;

-- name: GetFeedFollowsForUser :many
SELECT 
    feed_follows.*,
    users.name as user_name,
    feeds.name as feed_name
FROM feed_follows
INNER JOIN feeds ON feeds.id = feed_follows.feed_id
INNER JOIN users ON users.id = feed_follows.user_id
WHERE feed_follows.user_id = $1;