CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users_permissions (
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES
    ('accounts:read'),
    ('accounts:create'),
    ('accounts:update'),
    ('accounts:delete'),
    ('accounts:deposit'),
    ('accounts:withdraw'),
    ('accounts:transfer');