-- name: CreateSession :one
INSERT INTO sessions (id, name, status) 
VALUES ($1, $2, $3) 
RETURNING id, name, status, created_at, updated_at;

-- name: GetSession :one
SELECT id, name, status, created_at, updated_at 
FROM sessions 
WHERE id = $1;

-- name: UpdateSessionStatus :exec
UPDATE sessions 
SET status = $2, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1;

-- name: UpdateSessionName :one
UPDATE sessions 
SET name = $2, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1
RETURNING id, name, status, created_at, updated_at;

-- name: DeleteSession :exec
DELETE FROM sessions 
WHERE id = $1;

-- name: ListSessions :many
SELECT id, name, status, created_at, updated_at 
FROM sessions 
ORDER BY created_at DESC;

-- name: CreateHighlight :one
INSERT INTO highlights (id, session_id, text, position) 
VALUES ($1, $2, $3, $4) 
RETURNING id, session_id, text, position, created_at;

-- name: GetHighlightsBySession :many
SELECT id, session_id, text, position, created_at 
FROM highlights 
WHERE session_id = $1 
ORDER BY position;

-- name: GetHighlight :one
SELECT id, session_id, text, position, created_at 
FROM highlights 
WHERE id = $1;

-- name: DeleteHighlightsBySession :exec
DELETE FROM highlights 
WHERE session_id = $1;

-- name: CreateInteraction :one
INSERT INTO interactions (id, highlight_id, question, answer) 
VALUES ($1, $2, $3, $4) 
RETURNING id, highlight_id, question, answer, created_at, updated_at;

-- name: UpdateInteractionAnswer :one
UPDATE interactions 
SET answer = $2, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1
RETURNING id, highlight_id, question, answer, created_at, updated_at;

-- name: UpdateInteractionQuestion :one
UPDATE interactions 
SET question = $2, updated_at = CURRENT_TIMESTAMP 
WHERE id = $1
RETURNING id, highlight_id, question, answer, created_at, updated_at;

-- name: GetInteractionsByHighlight :many
SELECT id, highlight_id, question, answer, created_at, updated_at 
FROM interactions 
WHERE highlight_id = $1;

-- name: GetInteraction :one
SELECT id, highlight_id, question, answer, created_at, updated_at 
FROM interactions 
WHERE id = $1;

-- name: GetInteractionByHighlight :one
SELECT id, highlight_id, question, answer, created_at, updated_at 
FROM interactions 
WHERE highlight_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteInteractionsByHighlight :exec
DELETE FROM interactions 
WHERE highlight_id = $1;