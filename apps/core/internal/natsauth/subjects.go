package natsauth

import "fmt"

func WorkerPermissions(scope WorkerScope) Permissions {
	org := scope.OrganizationID
	wt := scope.WorkTypeID
	consumer := fmt.Sprintf("worker-%d", scope.WorkerID)
	return Permissions{
		PublishAllow: []string{
			fmt.Sprintf("result.org.%d.>", org),
			fmt.Sprintf("config.request.org.%d.work_type.%d.>", org, wt),
			fmt.Sprintf("$JS.API.CONSUMER.CREATE.TASKS.%s.task.org.%d.work_type.%d.>", consumer, org, wt),
			fmt.Sprintf("$JS.API.CONSUMER.MSG.NEXT.TASKS.%s", consumer),
			fmt.Sprintf("$JS.ACK.TASKS.%s.>", consumer),
			"_INBOX.>",
		},
		SubscribeAllow: []string{
			fmt.Sprintf("task.org.%d.work_type.%d.>", org, wt),
			fmt.Sprintf("config.org.%d.work_type.%d.>", org, wt),
			"_INBOX.>",
		},
	}
}
