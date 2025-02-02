package pipeline

type Step struct {
	Name string
}

type Status string

const (
	Wait   Status = "wait"
	Done   Status = "done"
	Work   Status = "work"
	Cancel Status = "cancel"
	Error  Status = "error"
)

var StepWaitQueue = Step{
	Name: "В очереди",
}

var StepWaitSendQueue = Step{
	Name: "Ожидает отправки",
}

var StepWork = Step{
	Name: "В работе у воркера",
}

var StepDone = Step{
	Name: "Выполнено",
}
