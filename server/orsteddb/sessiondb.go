package orsteddb

import (
	"errors"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	"orsted/protobuf/orstedrpc"
)

func RegisterSessionDB(s *orstedrpc.SessionReq) (*orstedrpc.Session, error) {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("SESSIONS"))
	if err != nil {
		return nil, err
	}

	sessionId, err := bkt.NextSequence()
	if err != nil {
		return nil, err
	}

	res := &orstedrpc.Session{Id: strconv.FormatUint(sessionId, 10), Os: s.Os, Hostname: s.Hostname, Ip: s.Ip, Integrity: s.Integrity, User: s.User, Status: "alive", Lastseen: time.Now().Unix(), Chain: s.Chain, Transport: s.Transport}

	if buf, err := proto.Marshal(res); err != nil {
		return nil, err
	} else if err := bkt.Put([]byte(strconv.FormatUint(sessionId, 10)), buf); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return res, nil
}

func ListSessionsDb() (*orstedrpc.SessionList, error) {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("SESSIONS"))
	if err != nil {
		return nil, err
	}

	c := b.Cursor()

	var data []*orstedrpc.Session
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var m orstedrpc.Session
		proto.Unmarshal(v, &m)
		data = append(data, &m)
	}
	res := &orstedrpc.SessionList{Sessions: data}
	return res, nil
}

func GetSessionById(beaconId string) (*orstedrpc.Session, error) {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("SESSIONS"))
	if err != nil {
		return nil, err
	}

	c := b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		var m orstedrpc.Session
		proto.Unmarshal(v, &m)
        if m.Id == beaconId {
            return &m, nil
        }
	}
	return nil, errors.New("Beacon Id Not Found")
}

func UpdatePol(beaconId string) error {
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("SESSIONS"))
	if err != nil {
		return err
	}

    var s orstedrpc.Session

    buf := bkt.Get([]byte(beaconId))
    err = proto.Unmarshal(buf, &s);
	if err != nil {
		return err
	} 

    s.Lastseen = time.Now().Unix()
    newbuf, err := proto.Marshal(&s)
    err = bkt.Put([]byte(s.Id), newbuf)
    if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
 
