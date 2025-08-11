CREATE TABLE IF NOT EXISTS user_invitations
(
    token      bytea PRIMARY KEY,
    user_id    bigint                      NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT now()
);