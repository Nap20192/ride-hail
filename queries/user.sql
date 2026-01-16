-- name: CreateUser :one
INSERT INTO
    users (
        email,
        role,
        status,
        password_hash,
        salt,
        attrs
    )
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at, email, role, status, password_hash, salt, attrs;

-- name: GetUserByEmail :one
SELECT
    id,
    created_at,
    updated_at,
    email,
    role,
    status,
    password_hash,
    salt,
    attrs
FROM users
WHERE email = $1;
