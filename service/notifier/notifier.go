package notifier

type ActionType int

const (
	ActionUnknown ActionType = iota
	ActionCreate
	ActionUpdate
	ActionDelete
)

type Action struct {
	Type  ActionType
	Key   string
	Value *string
}

type Notification struct {
	Action Action
}

type Notifier interface {
	Notify(notification Notification) error
}
