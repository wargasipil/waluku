package waluku

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/vmihailenco/msgpack/v5"
)

type TradeStore interface {
	Get(key string, data any) error
	Set(key string, data any) error
	Delete(key string) error
}

type tradeStoreImpl struct {
	db *badger.DB
}

// Delete implements TradeStore.
func (t *tradeStoreImpl) Delete(key string) error {
	err := t.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	return err
}

// Get implements TradeStore.
func (t *tradeStoreImpl) Get(key string, data any) error {

	err := t.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		err = msgpack.Unmarshal(val, data)
		return err
	})
	return err
}

// Set implements TradeStore.
func (t *tradeStoreImpl) Set(key string, data any) error {

	err := t.db.Update(func(txn *badger.Txn) error {
		raw, err := msgpack.Marshal(data)
		if err != nil {
			return err
		}
		return txn.Set([]byte(key), raw)
	})

	return err
}

func NewTradeStore(fname string) TradeStore {
	opts := badger.DefaultOptions(fname).WithLoggingLevel(badger.ERROR)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	return &tradeStoreImpl{
		db: db,
	}
}
