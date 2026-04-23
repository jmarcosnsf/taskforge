-- name: CreateTask :one
INSERT INTO tasks (title,description,team_id)
VALUES($1, $2, $3)
RETURNING *;

-- name: ListTasks :many
SELECT * FROM tasks 
WHERE team_id = $1;

-- name: GetTaskByID :one
SELECT * FROM tasks
WHERE id = $1;

-- name: UpdateTaskStatus :exec
UPDATE tasks SET status = $1, updated_at = NOW()
WHERE id = $2;

-- name: AssignTask :exec
UPDATE tasks SET user_id = $1, updated_at = NOW()
WHERE id = $2;

-- name: RemoveTask :exec
DELETE FROM tasks WHERE id = $1;