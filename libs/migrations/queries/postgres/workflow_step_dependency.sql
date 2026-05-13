-- name: CreateWorkflowStepDependency :exec
INSERT INTO "workflow_step_dependency" (step_id, depends_on_step_id, outcome, output_index)
VALUES ($1, $2, $3, $4);

-- name: ListWorkflowStepDependencies :many
SELECT * FROM "workflow_step_dependency"
WHERE step_id = $1
ORDER BY depends_on_step_id;

-- name: ListDependenciesByVersionID :many
SELECT wsd.*
FROM "workflow_step_dependency" wsd
JOIN "workflow_step" target_step ON target_step.id = wsd.step_id
JOIN "workflow_step" source_step ON source_step.id = wsd.depends_on_step_id
WHERE target_step.workflow_version_id = $1
  AND source_step.workflow_version_id = $1
  AND target_step.deleted_at IS NULL
  AND source_step.deleted_at IS NULL
ORDER BY wsd.step_id, wsd.depends_on_step_id;

-- name: DeleteWorkflowStepDependency :exec
DELETE FROM "workflow_step_dependency"
WHERE step_id = $1 AND depends_on_step_id = $2;

-- name: DeleteDependenciesByStepID :exec
DELETE FROM "workflow_step_dependency"
WHERE step_id = $1 OR depends_on_step_id = $1;
