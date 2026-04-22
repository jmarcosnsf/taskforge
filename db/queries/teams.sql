-- name: CreateTeam :one
INSERT INTO teams (name,description, owner_id)
VALUES ($1,$2,$3)
RETURNING *;

-- name: GetTeamByID :one
SELECT * FROM teams WHERE id = $1;

