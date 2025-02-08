-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
           $1,
           $2,
           $3,
           $4
       )
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1 LIMIT 1;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: DeleteUser :exec
DELETE FROM users
 WHERE id = $1;

-- name: DeleteUserByName :exec
DELETE FROM users
WHERE name = $1;