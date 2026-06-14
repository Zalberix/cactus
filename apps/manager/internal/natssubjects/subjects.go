package natssubjects

import "fmt"

type WorkerScope struct {
	OrganizationID int32
	WorkTypeID     int32
}

func Task(scope WorkerScope, schemaID, revisionID, workflowRunID int32) string {
	return fmt.Sprintf(
		"task.org.%d.work_type.%d.schema.%d.revision.%d.%d",
		scope.OrganizationID,
		scope.WorkTypeID,
		schemaID,
		revisionID,
		workflowRunID,
	)
}

func TaskFilter(scope WorkerScope) string {
	return fmt.Sprintf("task.org.%d.work_type.%d.>", scope.OrganizationID, scope.WorkTypeID)
}

func Result(organizationID, workflowRunID, stepID int32) string {
	return fmt.Sprintf("result.org.%d.%d.%d", organizationID, workflowRunID, stepID)
}

func Config(organizationID, workTypeID, revisionID int32) string {
	return fmt.Sprintf("config.org.%d.work_type.%d.revision.%d", organizationID, workTypeID, revisionID)
}

func ConfigRequest(organizationID, workTypeID, revisionID int32) string {
	return fmt.Sprintf("config.request.org.%d.work_type.%d.revision.%d", organizationID, workTypeID, revisionID)
}
