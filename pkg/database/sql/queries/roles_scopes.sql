-- name: AssignScopeToRole :exec
INSERT INTO roles_scopes (role_id, scope_id) 
VALUES ($1, $2) 
ON CONFLICT DO NOTHING;

-- name: RemoveScopeFromRole :exec
DELETE FROM roles_scopes 
WHERE role_id = $1 AND scope_id = $2;

-- name: GetScopesForRole :many
SELECT s.id, s.name, s.description 
FROM scopes s
INNER JOIN roles_scopes rs ON s.id = rs.scope_id
WHERE rs.role_id = $1
ORDER BY s.name;

-- name: GetRolesWithScope :many
SELECT r.id, r.name, r.description 
FROM roles r
INNER JOIN roles_scopes rs ON r.id = rs.role_id
WHERE rs.scope_id = $1
ORDER BY r.name;

-- name: RemoveAllScopesFromRole :exec
DELETE FROM roles_scopes WHERE role_id = $1;

-- name: RoleHasScope :one
SELECT EXISTS(SELECT 1 FROM roles_scopes WHERE role_id = $1 AND scope_id = $2);