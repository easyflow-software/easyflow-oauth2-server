-- Drop triggers first (before dropping tables they reference)
DROP TRIGGER IF EXISTS update_oauth_clients_updated_at ON oauth_clients;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_scopes_updated_at ON scopes;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS users_roles;
DROP TABLE IF EXISTS roles_scopes;
DROP TABLE IF EXISTS oauth_clients_scopes;
DROP TABLE IF EXISTS oauth_clients;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS scopes;

-- Drop the function last (since triggers reference it)
DROP FUNCTION IF EXISTS trigger_updated_at();