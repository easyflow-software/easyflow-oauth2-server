-- name: CreateOAuthClient :one
INSERT INTO oauth_clients (client_id, client_secret_hash, name, description, redirect_uris, grant_types)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, client_id, name, description, redirect_uris, grant_types, created_at, updated_at;

-- name: GetOAuthClient :one
SELECT id, client_id, client_secret_hash, name, description, redirect_uris, grant_types, created_at, updated_at, authorization_code_valid_duration, access_token_valid_duration, refresh_token_valid_duration, access_token_valid_duration, refresh_token_valid_duration
FROM oauth_clients
WHERE id = $1;

-- name: GetOAuthClientByClientID :one
SELECT
    oc.id,
    oc.client_id,
    oc.client_secret_hash,
    oc.name,
    oc.description,
    oc.redirect_uris,
    oc.grant_types,
    oc.created_at,
    oc.updated_at,
    oc.authorization_code_valid_duration,
    oc.access_token_valid_duration,
    oc.refresh_token_valid_duration,
    COALESCE(ARRAY_AGG(DISTINCT(s.name)) FILTER (WHERE s.name IS NOT NULL), ARRAY[]::TEXT[])::TEXT[] as scopes
FROM oauth_clients oc
LEFT JOIN oauth_clients_scopes ocs ON oc.id = ocs.oauth_client_id
LEFT JOIN scopes s ON ocs.scope_id = s.id
WHERE client_id = $1
GROUP BY
    oc.id,
    oc.client_id,
    oc.client_secret_hash,
    oc.name,
    oc.description,
    oc.redirect_uris,
    oc.grant_types,
    oc.created_at,
    oc.updated_at,
    oc.authorization_code_valid_duration,
    oc.access_token_valid_duration,
    oc.refresh_token_valid_duration;

-- name: ListOAuthClients :many
SELECT id, client_id, name, description, redirect_uris, grant_types, created_at, updated_at
FROM oauth_clients
ORDER BY name;

-- name: UpdateOAuthClient :one
UPDATE oauth_clients
SET name = $2, description = $3, redirect_uris = $4, grant_types = $5
WHERE id = $1
RETURNING id, client_id, name, description, redirect_uris, grant_types, created_at, updated_at;

-- name: UpdateOAuthClientSecret :exec
UPDATE oauth_clients
SET client_secret_hash = $2
WHERE id = $1;

-- name: DeleteOAuthClient :exec
DELETE FROM oauth_clients WHERE id = $1;

-- name: ClientIDExists :one
SELECT EXISTS(SELECT 1 FROM oauth_clients WHERE client_id = $1);

-- name: ValidateRedirectURI :one
SELECT EXISTS(SELECT 1 FROM oauth_clients WHERE client_id = $1 AND $2 = ANY(redirect_uris));
