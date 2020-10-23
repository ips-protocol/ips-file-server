package jobs

type Job interface {
	Enqueue(taskTime int64)
	Dequeue() (*Job, error)
}
