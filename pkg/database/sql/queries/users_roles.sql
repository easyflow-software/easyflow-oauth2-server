-- name: AssignRoleToUser :exec
INSERT INTO users_roles (user_id, role_id) 
VALUES ($1, $2) 
ON CONFLICT DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM users_roles 
WHERE user_id = $1 AND role_id = $2;

-- name: GetUserRoles :many
SELECT r.id, r.name, r.description 
FROM roles r
INNER JOIN users_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1
ORDER BY r.name;

-- name: GetUsersWithRole :many
SELECT u.id, u.email, u.first_name, u.last_name 
FROM users u
INNER JOIN users_roles ur ON u.id = ur.user_id
WHERE ur.role_id = $1
ORDER BY u.email;

-- name: RemoveAllRolesFromUser :exec
DELETE FROM users_roles WHERE user_id = $1;

-- name: UserHasRole :one
SELECT EXISTS(SELECT 1 FROM users_roles WHERE user_id = $1 AND role_id = $2);

-- name: GetUserScopes :many
SELECT DISTINCT s.name 
FROM scopes s
INNER JOIN roles_scopes rs ON s.id = rs.scope_id
INNER JOIN users_roles ur ON rs.role_id = ur.role_id
WHERE ur.user_id = $1
ORDER BY s.name;