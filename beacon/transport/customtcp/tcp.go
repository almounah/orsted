package customtcp
//Useless

import (
	"bytes"
	"net"
	"orsted/beacon/utils"
)

var TcpDelimiter []byte = []byte("dhddudoO\r\n")

func ReceiveData(conn net.Conn) []byte {
    defer conn.Close()

    var res []byte
    // Read data from connection and send response
    // Data is marked by End "oO\n"
    buf := make([]byte, 1024)
    for {
        _, err := conn.Read(buf)
        res = append(res, buf...)
        if err != nil {
            utils.Print("Error While Reading ", err.Error())
            return nil
        }
        if bytes.HasSuffix(res, TcpDelimiter) {
            break
        }
    }
    return res

}

func SendData(conn net.Conn, data []byte) {
    data = append(data, TcpDelimiter...)
    conn.Write(data)
}
