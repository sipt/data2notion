package notice

type INotifier interface {
	Title() string
	Send(message string, args ...interface{})
	SendError(err error)
	Update(message string, args ...interface{})
	Append(message string, args ...interface{})
	MakeThread(title string, args ...interface{}) INotifier
	Finish(successTag, failedTag string)
}
