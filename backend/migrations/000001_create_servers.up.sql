CREATE TABLE IF NOT EXISTS servers (
    id text PRIMARY KEY,
    name text NOT NULL,
    hostname text NOT NULL,
    ip text NOT NULL DEFAULT '',
    operating_system text NOT NULL DEFAULT '',
    status text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT servers_hostname_key UNIQUE (hostname)
);
