package orsteddb

import (
	"errors"

	"google.golang.org/protobuf/proto"

	"orsted/protobuf/orstedrpc"
)

func HostFileDb(h *orstedrpc.Host) (error) {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("HOSTED_FILES"))
	if err != nil {
		return err
	}

	if buf, err := proto.Marshal(h); err != nil {
		return err
	} else if err := bkt.Put([]byte(h.Filename), buf); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func GetFileDataDb(fileName string) (data []byte, err error) {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("HOSTED_FILES"))
	if err != nil {
		return nil, err
	}

	c := b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		var m orstedrpc.Host
		proto.Unmarshal(v, &m)
        if m.Filename == fileName {
            return m.GetData(), nil
        }
	}
	return nil, errors.New("File Name not found")
}

func ViewHostedFileDb() (*orstedrpc.HostList, error) {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("HOSTED_FILES"))
	if err != nil {
		return nil, err
	}

	c := b.Cursor()

	var data []*orstedrpc.Host
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var m orstedrpc.Host
		proto.Unmarshal(v, &m)
        data = append(data, &m)
	}
    res := &orstedrpc.HostList{Hostlist: data}
	return res, nil

}

func UnHostFileDb(h *orstedrpc.Host) error {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	b, err := tx.CreateBucketIfNotExists([]byte("HOSTED_FILES"))
	if err != nil {
		return err
	}
	err = b.Delete([]byte(h.Filename))
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

