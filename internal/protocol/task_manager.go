package protocol

type Task struct {
	Command         Array
	ResponseChannel chan DataType
}

const (
	taskQueueSize = 1000
)

var tasks chan Task

func AppendTask(task Task) {
	tasks <- task
}

func Start() {
	tasks = make(chan Task, taskQueueSize)
	for task := range tasks {
		response := processCommand(task.Command)
		task.ResponseChannel <- response
	}
}
