package orsteddb

import (
	"errors"

	"google.golang.org/protobuf/proto"

	"orsted/protobuf/orstedrpc"
	"orsted/server/utils"
)

func HostFileDb(h *orstedrpc.Host) error {
	// Adding File to DB and Getting hash
	hashData, err := AddFileToDb(h.Data)
	if err != nil {
		return err
	}

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

	// Putting Hash instead of Byte
	var newH orstedrpc.Host
	newH.Filename = h.Filename
	newH.Data = hashData

	// Marhselling new host with hash
	buf, err := proto.Marshal(&newH)
	if err != nil {
		return err
	}

	err = bkt.Put([]byte(h.Filename), buf)
	if err != nil {
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
			res, err := ConvertHostDataHashToByte(db, &m)
			if err != nil {
				utils.PrintDebug("Error Converting Host Data to Hash", err)
				return nil, err
			}
			return res.GetData(), nil
		}
	}
	return nil, errors.New("File Name not found")
}

func ViewHostedFileDb() (*orstedrpc.HostList, error) {
	// No need to convert to hash as we only use this function to print the filename
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
