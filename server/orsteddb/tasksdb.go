package orsteddb

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"orsted/protobuf/orstedrpc"
	"orsted/server/utils"
)

func AddTaskDb(treq *orstedrpc.TaskReq) (*orstedrpc.Task, error) {
	utils.PrintDebug("Adding Task")
	// Hash request Data to avoid storing it multiple times in DB
	h, err := AddFileToDb(treq.Reqdata)
	if err != nil {
		return nil, err
	}
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("TASKS"))
	if err != nil {
		return nil, err
	}

	taskId, err := bkt.NextSequence()
	if err != nil {
		return nil, err
	}

	res := &orstedrpc.Task{TaskId: strconv.FormatUint(taskId, 10), 
                           BeacondId: treq.BeacondId,
                           State: "pending",       
                           Command: treq.Command,
                           PrettyCommand: treq.PrettyCommand,
                           Reqdata: h,
                           SentAt: time.Now().Unix()}

	if buf, err := proto.Marshal(res); err != nil {
		return nil, err
	} else if err := bkt.Put([]byte(strconv.FormatUint(taskId, 10)), buf); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return res, nil
}

func ChangeTaskState(taskId string, state string) error {
	utils.PrintDebug("Changing Task State")
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("TASKS"))
	if err != nil {
		return err
	}

    var t orstedrpc.Task

    buf := bkt.Get([]byte(taskId))
    err = proto.Unmarshal(buf, &t);
	if err != nil {
		return err
	} 

    t.State = state
    newbuf, err := proto.Marshal(&t)
    err = bkt.Put([]byte(t.TaskId), newbuf)
    if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func ListTasksDb(beaconId string, states []string) (*orstedrpc.TaskList, error) {
	utils.PrintDebug("Listing Task")
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("TASKS"))
	if err != nil {
		return nil, err
	}

	c := b.Cursor()

	var data []*orstedrpc.Task
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var m orstedrpc.Task
		proto.Unmarshal(v, &m)
        if m.BeacondId == beaconId && slices.Contains(states, m.State) {
			converted, err := ConvertTaskDataHashToByte(db, &m)
			if err == nil {
				data = append(data, converted)
			}
        }
	}
    res := &orstedrpc.TaskList{BeaconId: beaconId, Tasks: data}
	return res, nil

}

func GetTaskState(taskId string) (string, error) {
	utils.PrintDebug("Getting Task State")
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("TASKS"))
	if err != nil {
		return "", err
	}

    var t orstedrpc.Task

    buf := bkt.Get([]byte(taskId))
    err = proto.Unmarshal(buf, &t);
	if err != nil {
		return "", err
	} 

    return t.State, nil
}

func SetTaskResponse(trep *orstedrpc.TaskRep) (*orstedrpc.Task, error) {
	utils.PrintDebug("Setting Task Response")
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bkt, err := tx.CreateBucketIfNotExists([]byte("TASKS"))
	if err != nil {
		return nil, err
	}

    var t orstedrpc.Task

    buf := bkt.Get([]byte(trep.TaskId))
    err = proto.Unmarshal(buf, &t);
	if err != nil {
		return nil, err
	} 
 
    // TODO: Fix this because malicious actor can send ongoing bytes
    if t.State != "sent" && !strings.Contains(t.State, "ongoing") {
		return nil, errors.New("Cannot send result for task not in 'sent' state. Something Phishing is Going on ...")
    }
    t.State = trep.State
    t.Response = trep.Response
    newbuf, err := proto.Marshal(&t)
    err = bkt.Put([]byte(t.TaskId), newbuf)
    if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &t, nil
}

func GetTaskByIdDb(taskId string) (*orstedrpc.Task, error) {
	utils.PrintDebug("Getting Task by ID")
	db := Initialise()
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("TASKS"))
	if err != nil {
		return nil, err
	}

	c := b.Cursor()

	for k, v := c.First(); k != nil; k, v = c.Next() {
		var m orstedrpc.Task
		proto.Unmarshal(v, &m)
        if m.TaskId == taskId {
			t, err := ConvertTaskDataHashToByte(db, &m)
			if err != nil {
				return nil, errors.New("Beacon Id Not Found")
			}
            return t, nil
        }
	}
	return nil, errors.New("Beacon Id Not Found")
}
