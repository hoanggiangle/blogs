package sdwp

type Task struct {
	id      string
	handler func() error
}

// Create a new task for worker pool
//   * id string: task identify
//   * hdl: a function to handle task
func NewTask(id string, hdl func() error) Task {
	return Task{
		id:      id,
		handler: hdl,
	}
}

func (t Task) GetId() string {
	return t.id
}
