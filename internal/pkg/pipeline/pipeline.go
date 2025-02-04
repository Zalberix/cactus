package pipeline

type Step struct {
	Step int32  `json:"step"`
	Name string `json:"name"`
}

type Status string

const (
	Wait   Status = "wait"
	Done   Status = "done"
	Work   Status = "work"
	Cancel Status = "cancel"
	Error  Status = "error"
)
