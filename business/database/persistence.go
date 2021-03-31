package database

import (
	"time"

	"github.com/etcd-io/bbolt"
)

type Data struct {
	id        string
	Indexes   []string
	TimeStamp time.Time
	Payload   []byte
}

func OpenDB(path string) (*bbolt.DB, error) {
	db, err := bbolt.Open(path, 0664, bbolt.DefaultOptions)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Put(db *bbolt.DB, databaseName string, id string, indexes []string, timestamp *time.Time, data []byte) error {

	//TODO: id have timestamp?
	tt := time.Now()
	if timestamp != nil {
		tt = *timestamp
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		bkActor, err := tx.CreateBucketIfNotExists([]byte(databaseName))
		if err != nil {
			return bbolt.ErrBucketNotFound
		}
		dayBucket := tt.Format("2006-01")
		bkDayBucket, err := bkActor.CreateBucketIfNotExists([]byte(dayBucket))
		if err != nil {
			return bbolt.ErrBucketNotFound
		}
		if err := bkDayBucket.Put([]byte(id), []byte(id)); err != nil {
			return err
		}
		for _, index := range indexes {
			bk, err := bkDayBucket.CreateBucketIfNotExists([]byte(index))
			if err != nil {
				return bbolt.ErrBucketNotFound
			}
			if err := bk.Put([]byte(id), []byte(id)); err != nil {
				return err
			}
		}
		if err := bkDayBucket.Put([]byte(id), data); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func Get(db *bbolt.DB, databaseName string, id string) ([]byte, error) {

	funcGetDateDatabase := func(id string) []byte {
		return nil
	}

	result := make([]byte, 0)

	if bkName := funcGetDateDatabase(id); bkName != nil {
		if err := db.View(func(tx *bbolt.Tx) error {
			bk := tx.Bucket([]byte(databaseName))
			if bk == nil {
				return bbolt.ErrBucketNotFound
			}
			bkDate := bk.Bucket(bkName)
			if bk == nil {
				return bbolt.ErrBucketNotFound
			}
			data := bkDate.Get([]byte(id))
			if data == nil {
				return nil
			}
			result = append(result, data...)
			return nil
		}); err != nil {
			return nil, err
		}
		if len(result) <= 0 {
			return nil, nil
		}
		return result, nil
	}
	if err := db.View(func(tx *bbolt.Tx) error {
		bk := tx.Bucket([]byte(databaseName))
		if bk == nil {
			return bbolt.ErrBucketNotFound
		}

		funcGetData := func(k []byte) ([]byte, error) {

			bkTemp := bk.Bucket(k)
			if bkTemp == nil {
				return nil, nil
			}
			data := bkTemp.Get([]byte(id))
			if data == nil {
				return nil, nil
			}
			return data, nil
		}

		it := bk.Cursor()
		kFirst, vFirst := it.First()
		if vFirst == nil {
			if data, err := funcGetData(kFirst); err == nil && data != nil {
				result = append(result, data...)
				return nil
			}
		}
		for {
			if k, v := it.Next(); v == nil && k == nil {
				return nil
			} else if v == nil {

				if data, err := funcGetData(k); err == nil && data != nil {
					result = append(result, data...)
					return nil
				}
			} else {
				continue
			}
		}
	}); err != nil {
		return nil, err
	}
	if len(result) <= 0 {
		return nil, nil
	}
	return result, nil
}

type QueryFilter struct {
	Indexes    []string
	BeforeFrom *time.Time
	AfterFrom  *time.Time
}

func Query(db *bbolt.DB, databaseName string, queryFilter *QueryFilter, funcFilter func(v interface{})) ([]byte, error) {

	// var d1 []byte
	// var d2 []byte
	// if queryFilter.BeforeFrom != nil {
	// d1 = []byte(queryFilter.BeforeFrom.Format("2006-01"))
	// }

	return nil, nil
}
