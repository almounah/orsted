package pivot

import (
	"net"
	"orsted/beacon/utils"
	"strconv"
)

type TcpStuff struct {
	Id   string
	Ip   string
	Port string
	L    net.Listener
}

func NewTcpStuff(address string) (TcpStuff, error) {
	  host, port, err := net.SplitHostPort(address)
    if err != nil {
        utils.Print("Error parsing address:", err)
        return TcpStuff{}, err
    }

	var t TcpStuff
	t.Ip = host
	t.Port = port
	Pivox++
	t.Id = strconv.Itoa(Pivox)
	return t, nil

}

func (t *TcpStuff) Start() error {
	listener, err := net.Listen("tcp", t.Ip+":"+t.Port)
	if err != nil {
		utils.Print("Error Listening", err.Error())
		return err
	}

	t.L = listener

	for {
		conn, err := listener.Accept()
		if err != nil {
			utils.Print("Error Accepting Connection", err.Error())
			return err
		}

		// Handle incoming connection here
		go handleConnection(conn)
	}
}

func (t *TcpStuff) Stop()  {
	if t.L != nil {

		t.L.Close()
	}
	
}

func (t *TcpStuff) GetString() string {
	return t.Ip + ":" + t.Port

}

func (t *TcpStuff) GetID() string {
	return t.Id

}

func (t *TcpStuff) GetType() string {
	return "tcp"

}
