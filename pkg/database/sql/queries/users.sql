-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name) 
VALUES ($1, $2, $3, $4) 
RETURNING id, email, first_name, last_name, created_at, updated_at;

-- name: GetUser :one
SELECT id, email, first_name, last_name, created_at, updated_at 
FROM users 
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT 
    u.id, 
    u.email, 
    u.password_hash, 
    u.first_name, 
    u.last_name, 
    u.created_at, 
    u.updated_at,
    ARRAY_AGG(s.name)::TEXT[] as scopes
FROM users u
JOIN users_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
JOIN roles_scopes rs ON r.id = rs.role_id
JOIN scopes s ON rs.scope_id = s.id
WHERE u.email = $1
GROUP BY u.id, u.email, u.password_hash, u.first_name, u.last_name, u.created_at, u.updated_at;

-- name: ListUsers :many
SELECT id, email, first_name, last_name, created_at, updated_at 
FROM users 
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users 
SET email = $2, first_name = $3, last_name = $4 
WHERE id = $1 
RETURNING id, email, first_name, last_name, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users 
SET password_hash = $2 
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: GetUserWithRolesAndScopes :one
SELECT u.id, u.email, u.first_name, u.last_name, u.created_at, u.updated_at,
       COALESCE(array_agg(DISTINCT r.name) FILTER (WHERE r.name IS NOT NULL), ARRAY[]::TEXT[]) as role_names,
       COALESCE(array_agg(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), ARRAY[]::TEXT[]) as scope_names
FROM users u
LEFT JOIN users_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
LEFT JOIN roles_scopes rs ON r.id = rs.role_id
LEFT JOIN scopes s ON rs.scope_id = s.id
WHERE u.id = $1
GROUP BY u.id, u.email, u.first_name, u.last_name, u.created_at, u.updated_at;