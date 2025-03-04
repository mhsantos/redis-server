package taskmanager

import (
	"github.com/mhsantos/redis-server/internal/commands"
	"github.com/mhsantos/redis-server/internal/protocol"
)

type Task struct {
	Command         protocol.Array
	ResponseChannel chan protocol.DataType
	ErrorChannel    chan error
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
		response := commands.ProcessCommand(task.Command)
		task.ResponseChannel <- response
	}
}
