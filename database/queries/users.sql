-- name: CreateUser :one
INSERT INTO users 
(username,email,password,full_name) VALUES (
    $1,$2,$3,$4
) RETURNING *;


-- name: FindUserByName :one
SELECT * FROM users
WHERE username=$1 LIMIT 1;
