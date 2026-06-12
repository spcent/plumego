package app

import (
	"context"
	"errors"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

// kvAdapter narrows *kvstore.KVStore to the session.KV contract, translating
// the store's not-found/expired errors into the (found=false, nil) shape.
type kvAdapter struct {
	kv *kvstore.KVStore
}

func (a kvAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return a.kv.SetContext(ctx, key, value, ttl)
}

func (a kvAdapter) Get(ctx context.Context, key string) ([]byte, bool, error) {
	v, err := a.kv.GetContext(ctx, key)
	if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

func (a kvAdapter) Delete(ctx context.Context, key string) error {
	err := a.kv.DeleteContext(ctx, key)
	if errors.Is(err, kvstore.ErrKeyNotFound) {
		return nil
	}
	return err
}
