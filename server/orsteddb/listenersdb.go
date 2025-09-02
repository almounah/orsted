package orsteddb

import (
	"strconv"

	"google.golang.org/protobuf/proto"

	"orsted/protobuf/orstedrpc"
)

func AddListener(req *orstedrpc.ListenerReq) (*orstedrpc.ListenerJob, error) {
	db := Initialise()
	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("LISTENERS"))
	if err != nil {
		return nil, err
	}
	listenerId, err := bkt.NextSequence()
	if err != nil {
		return nil, err
	}

	res := &orstedrpc.ListenerJob{Id: strconv.FormatUint(listenerId, 10), Ip: req.Ip, Port: req.Port, ListenerType: req.ListenerType, CertPath: req.CertPath, KeyPath: req.KeyPath}
	if buf, err := proto.Marshal(res); err != nil {
		return nil, err
	} else if err := bkt.Put([]byte(strconv.FormatUint(listenerId, 10)), buf); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return res, nil
}

func ListListener() (*orstedrpc.ListenerJobList, error) {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("LISTENERS"))
	if err != nil {
		return nil, err
	}

	c := b.Cursor()

	var data []*orstedrpc.ListenerJob
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var m orstedrpc.ListenerJob
		proto.Unmarshal(v, &m)
		data = append(data, &m)
	}
	res := &orstedrpc.ListenerJobList{Listener: data}
	return res, nil
}

func DeleteListener(listener *orstedrpc.ListenerJob) error {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	b, err := tx.CreateBucketIfNotExists([]byte("LISTENERS"))
	if err != nil {
		return err
	}
	err = b.Delete([]byte(listener.Id))
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
