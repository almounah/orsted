package gowinrm

import (

	"github.com/gofrs/uuid"
)

func genUUID() string {
	id := uuid.Must(uuid.NewV4())
	return "uuid:" + id.String()
}


// NewDeleteShellRequest ...
func NewDeleteShellRequest(uri, shellID string, params *Parameters) string {
	message := "TODO"

	return message
}


//NewSendInputRequest NewSendInputRequest
func NewSendInputRequest(uri, shellID, commandID string, input []byte, eof bool, params *Parameters) string{
	message := "NOT IMPLEMENTED"
	return message
}

//NewSignalRequest NewSignalRequest
func NewSignalRequest(uri string, shellID string, commandID string, params *Parameters) string {
	message := "NOT IMPLEMENTED"
	return message
}
