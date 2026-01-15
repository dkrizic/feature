package notifier

import (
	"context"
)

type ActionType string

const (
	ActionUnknown ActionType = "unknown"
	ActionCreate  ActionType = "create"
	ActionUpdate  ActionType = "update"
	ActionDelete  ActionType = "delete"
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
	Notify(ctx context.Context, notification Notification) error
}

func CreateNotifucation(key string, value string) Notification {
	return Notification{
		Action: Action{
			Type:  ActionCreate,
			Key:   key,
			Value: &value,
		},
	}
}

func UpdateNotification(key string, value string) Notification {
	return Notification{
		Action: Action{
			Type:  ActionUpdate,
			Key:   key,
			Value: &value,
		},
	}
}

func DeleteNotification(key string) Notification {
	return Notification{
		Action: Action{
			Type: ActionDelete,
			Key:  key,
		},
	}
}
