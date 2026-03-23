-- name: CreateWorkflowStepDependency :exec
INSERT INTO "workflow_step_dependency" (step_id, depends_on_step_id, outcome)
VALUES ($1, $2, $3);

-- name: ListWorkflowStepDependencies :many
SELECT * FROM "workflow_step_dependency"
WHERE step_id = $1
ORDER BY depends_on_step_id;

-- name: ListDependenciesByVersionID :many
SELECT wsd.*
FROM "workflow_step_dependency" wsd
JOIN "workflow_step" ws ON ws.id = wsd.step_id
WHERE ws.workflow_version_id = $1 AND ws.deleted_at IS NULL
ORDER BY wsd.step_id, wsd.depends_on_step_id;

-- name: DeleteWorkflowStepDependency :exec
DELETE FROM "workflow_step_dependency"
WHERE step_id = $1 AND depends_on_step_id = $2;

-- name: DeleteDependenciesByStepID :exec
DELETE FROM "workflow_step_dependency"
WHERE step_id = $1;
