package natsauth

import "fmt"

func taskFilterSubject(organizationID, workTypeID, settingsSchemaID int32) string {
	return fmt.Sprintf("task.org.%d.work_type.%d.schema.%d.>", organizationID, workTypeID, settingsSchemaID)
}

func taskConsumerName(organizationID, workTypeID, settingsSchemaID int32) string {
	return fmt.Sprintf("task-org-%d-work-type-%d-schema-%d", organizationID, workTypeID, settingsSchemaID)
}

func WorkerPermissions(scope WorkerScope) Permissions {
	org := scope.OrganizationID
	wt := scope.WorkTypeID
	schema := scope.WorkerSettingsSchemaID
	consumer := taskConsumerName(org, wt, schema)
	return Permissions{
		PublishAllow: []string{
			fmt.Sprintf("result.org.%d.>", org),
			fmt.Sprintf("config.request.org.%d.work_type.%d.>", org, wt),
			"$JS.API.STREAM.INFO.CONFIGS",
			fmt.Sprintf("$JS.API.DIRECT.GET.CONFIGS.config.org.%d.work_type.%d.>", org, wt),
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.TASKS.%s.%s", consumer, taskFilterSubject(org, wt, schema)),
			fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.TASKS.%s", consumer),
			fmt.Sprintf("$JS.ACK.TASKS.%s.>", consumer),
			"_INBOX.>",
		},
		SubscribeAllow: []string{
			taskFilterSubject(org, wt, schema),
			fmt.Sprintf("config.org.%d.work_type.%d.>", org, wt),
			"_INBOX.>",
		},
	}
}
