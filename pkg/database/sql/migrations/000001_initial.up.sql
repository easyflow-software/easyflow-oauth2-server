CREATE OR REPLACE FUNCTION trigger_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE scopes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    description TEXT
);

CREATE TRIGGER update_scopes_updated_at
    BEFORE UPDATE
    ON
        scopes
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_updated_at();

CREATE TABLE roles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    name TEXT UNIQUE NOT NULL,
    description TEXT
);

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE
    ON
        roles
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_updated_at();

CREATE TABLE roles_scopes (
    role_id uuid REFERENCES roles(id) ON DELETE CASCADE,
    scope_id uuid REFERENCES scopes(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, scope_id)
);

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    first_name TEXT, -- optional used for display purposes
    last_name TEXT -- optional used for display purposes
);

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON
        users
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_updated_at();

CREATE TABLE users_roles (
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    role_id uuid REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TYPE grant_types AS ENUM ('code', 'refresh_token', 'client_credentials', 'device_code', 'pkce');

CREATE TABLE oauth_clients (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    client_id TEXT UNIQUE NOT NULL,
    client_secret_hash TEXT, -- NULL for public clients (PKCE)
    name TEXT NOT NULL,
    description TEXT,
    redirect_uris TEXT[] NOT NULL, -- Array of allowed redirect URIs
    grant_types grant_types[] DEFAULT ARRAY['code'],
    is_public BOOLEAN DEFAULT FALSE, -- True for PKCE clients
    token_endpoint_auth_method TEXT DEFAULT 'client_secret_basic'
);

CREATE TRIGGER update_oauth_clients_updated_at
    BEFORE UPDATE ON oauth_clients
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_updated_at();


CREATE TABLE oauth_clients_scopes (
    oauth_client_id uuid REFERENCES oauth_clients(id) ON DELETE CASCADE,
    scope_id uuid REFERENCES scopes(id) ON DELETE CASCADE,
    PRIMARY KEY (oauth_client_id, scope_id)
);
