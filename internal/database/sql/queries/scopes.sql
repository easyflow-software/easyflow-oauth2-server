-- name: CreateScope :one
INSERT INTO scopes (name, description) 
VALUES ($1, $2) 
RETURNING id, name, description, created_at, updated_at;

-- name: GetScope :one
SELECT id, name, description, created_at, updated_at 
FROM scopes 
WHERE id = $1;

-- name: GetScopeByName :one
SELECT id, name, description, created_at, updated_at 
FROM scopes 
WHERE name = $1;

-- name: ListScopes :many
SELECT id, name, description, created_at, updated_at 
FROM scopes 
ORDER BY name;

-- name: UpdateScope :one
UPDATE scopes 
SET name = $2, description = $3 
WHERE id = $1 
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteScope :exec
DELETE FROM scopes WHERE id = $1;

-- name: ScopeExistsByName :one
SELECT EXISTS(SELECT 1 FROM scopes WHERE name = $1);