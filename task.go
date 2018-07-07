package noodles

type Task struct {
	QueueName string
	FuncName  string
	Args      []interface{}
}
