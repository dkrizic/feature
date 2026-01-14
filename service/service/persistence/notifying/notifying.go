package notifying

// implements the persistence interface and wraps another persistence to send notifications on changes

import (
	"github.com/dkrizic/feature/service/notifier"
	"github.com/dkrizic/feature/service/service/persistence"
)

type NotifyingPersistence struct {
	wrapped  persistence.Persistence
	notifier notifier.Notifier
}

func NewNotifyingPersistence(wrapped persistence.Persistence, notifier notifier.Notifier) *NotifyingPersistence {
	return &NotifyingPersistence{
		wrapped:  wrapped,
		notifier: notifier,
	}
}
