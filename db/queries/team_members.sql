-- name: GetTeamsByUser :many
SELECT t.* FROM teams t
JOIN team_members tm ON t.id = tm.team_id
WHERE tm.user_id = $1;

-- name: AddTeamMember :exec
INSERT INTO team_members (user_id,team_id)
VALUES ($1,$2);

-- name: RemoveTeamMember :exec
DELETE FROM team_members WHERE user_id = $1 AND team_id = $2;

-- name: GetTeamMembers :many
SELECT u.* FROM users u 
JOIN team_members tm ON u.id = tm.user_id 
WHERE tm.team_id = $1;