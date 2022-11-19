CREATE TABLE users (
    id serial PRIMARY KEY,
    email text,
    hashed_password varchar(60),
    created timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tarantulas (
    id serial PRIMARY KEY,
    species text NOT NULL,
    name text NOT NULL,
    feed_interval_days integer DEFAULT 0,
    notify bool NOT NULL DEFAULT FALSE,
    img_url text NOT NULL,
    created timestamptz NOT NULL DEFAULT now(),
    owner_id int NOT NULL REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE sessions (
    token text PRIMARY KEY,
    data bytea NOT NULL,
    expiry timestamptz NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

