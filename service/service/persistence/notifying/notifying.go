package notifying

// implements the persistence interface and wraps another persistence to send notifications on changes

import (
	"context"

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

func (p *NotifyingPersistence) GetAll(ctx context.Context) ([]persistence.KeyValue, error) {
	return p.wrapped.GetAll(ctx)
}

func (p *NotifyingPersistence) PreSet(ctx context.Context, kv persistence.KeyValue) error {
	return p.wrapped.PreSet(ctx, kv)
}

func (p *NotifyingPersistence) Set(ctx context.Context, kv persistence.KeyValue) error {
	err := p.wrapped.Set(ctx, kv)
	if err != nil {
		return err
	}

	notification := notifier.UpdateNotification(kv.Key, kv.Value)
	return p.notifier.Notify(ctx, notification)
}

func (p *NotifyingPersistence) Delete(ctx context.Context, key string) error {
	err := p.wrapped.Delete(ctx, key)
	if err != nil {
		return err
	}

	notification := notifier.DeleteNotification(key)
	return p.notifier.Notify(ctx, notification)
}

func (p *NotifyingPersistence) Get(ctx context.Context, key string) (persistence.KeyValue, error) {
	return p.wrapped.Get(ctx, key)
}

func (p *NotifyingPersistence) Count(ctx context.Context) (int, error) {
	return p.wrapped.Count(ctx)
}
