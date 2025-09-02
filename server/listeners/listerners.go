package listeners

import (
	"errors"
	"orsted/server/utils"
)

type Listeners interface {
	StartListener() error
	StopListener() error
    GetListenerId() string
}

var LISTENERS_LIST []Listeners

func DeleteListenerById(Id string) error {
    utils.PrintInfo("Deleting Listener, ", Id)
	for i := 0; i < len(LISTENERS_LIST); i++ {
		if LISTENERS_LIST[i].GetListenerId() == Id {
			LISTENERS_LIST[i].StopListener()
			LISTENERS_LIST = append(LISTENERS_LIST[:i], LISTENERS_LIST[i+1:]...)
			return nil
		}
	}
	return errors.New("ID Not Found")
}
