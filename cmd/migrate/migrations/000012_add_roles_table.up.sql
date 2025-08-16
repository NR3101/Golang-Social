CREATE TABLE IF NOT EXISTS roles
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(50) NOT NULL UNIQUE,
    level       INT         NOT NULL DEFAULT 0,
    description TEXT
);

INSERT INTO roles (name, level, description)
VALUES
    ('user', 1, 'Regular user can create posts and comments'),
    ('moderator', 2, 'Moderator can modify posts and comments'),
    ('admin', 3, 'Administrator with full access');
