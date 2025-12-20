package orsteddb

import (
	"crypto/md5"
	"fmt"
	"orsted/protobuf/orstedrpc"
	"orsted/server/utils"

	"github.com/boltdb/bolt"
)

func HashBytes(d []byte) [16]byte {
	fmt.Println("Hashing")
	res := md5.Sum(d)
	fmt.Println("Done")
	return res
}

func AddFileToDb(f []byte) (h []byte, err error) {
	utils.PrintDebug("Adding file to db")
	db := Initialise()
	utils.PrintDebug("db initialised")
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("FILES"))
	if err != nil {
		return nil, err
	}


	hash := HashBytes(f)
	err = bkt.Put(hash[:], f)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return hash[:], nil
}

func RetrieveFileFromOpenDb(db *bolt.DB, hash []byte) ([]byte, error) {
	utils.PrintDebug("Retrieve task file from db")

    var valCopy []byte

    err := db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte("FILES"))
        if b == nil {
            return fmt.Errorf("bucket FILES not found")
        }

        v := b.Get(hash)
        if v == nil {
            return nil // not found
        }

        // COPY is mandatory
        valCopy = make([]byte, len(v))
        copy(valCopy, v)

        return nil
    })

    return valCopy, err
}

func ConvertTaskDataHashToByte(db *bolt.DB, taskWithHash *orstedrpc.Task) (*orstedrpc.Task, error) {
	utils.PrintDebug("Converting Task Data Hash")
	if taskWithHash == nil {
		return nil, fmt.Errorf("Nil task as input")
	}

	hash := taskWithHash.Reqdata
	f, err := RetrieveFileFromOpenDb(db, hash)
	if err != nil {
        return nil, err
	}
	var task orstedrpc.Task
	task.BeacondId = taskWithHash.BeacondId
	task.Command = taskWithHash.Command
	task.PrettyCommand = taskWithHash.PrettyCommand
	task.SentAt = taskWithHash.SentAt
	task.Response = taskWithHash.Response
	task.Reqdata = f
	task.State = taskWithHash.State
	task.TaskId = taskWithHash.TaskId

	return &task, nil
}
