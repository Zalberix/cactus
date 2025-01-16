package pipeline

type PipelineStep struct {
	Name string
}

type PipelineStatus string

const (
	Wait   PipelineStatus = "wait"
	Done   PipelineStatus = "done"
	Work   PipelineStatus = "work"
	Cancel PipelineStatus = "cancel"
	Error  PipelineStatus = "error"
)

var PipelineStepWaitQueue = PipelineStep{
	Name: "В очереди",
}

var PipelineStepWaitSendQueue = PipelineStep{
	Name: "Ожидает отправки",
}

var PipelineStepWork = PipelineStep{
	Name: "В работе у воркера",
}

var PipelineStepDone = PipelineStep{
	Name: "Выполнено",
}
