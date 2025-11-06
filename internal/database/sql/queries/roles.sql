-- name: CreateRole :one
INSERT INTO roles (name, description) 
VALUES ($1, $2) 
RETURNING id, name, description, created_at, updated_at;

-- name: GetRole :one
SELECT id, name, description, created_at, updated_at 
FROM roles 
WHERE id = $1;

-- name: GetRoleByName :one
SELECT id, name, description, created_at, updated_at 
FROM roles 
WHERE name = $1;

-- name: ListRoles :many
SELECT id, name, description, created_at, updated_at 
FROM roles 
ORDER BY name;

-- name: UpdateRole :one
UPDATE roles 
SET name = $2, description = $3 
WHERE id = $1 
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- name: GetRoleWithScopes :one
SELECT r.id, r.name, r.description, r.created_at, r.updated_at,
       COALESCE(array_agg(s.name) FILTER (WHERE s.name IS NOT NULL), ARRAY[]::TEXT[]) as scope_names
FROM roles r
LEFT JOIN roles_scopes rs ON r.id = rs.role_id
LEFT JOIN scopes s ON rs.scope_id = s.id
WHERE r.id = $1
GROUP BY r.id, r.name, r.description, r.created_at, r.updated_at;