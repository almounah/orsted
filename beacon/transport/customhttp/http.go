package customhttp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"orsted/beacon/utils"
	"orsted/profiles"
)

func SendRequest(conf profiles.ProfileConfig, endpointKey string, body []byte) []byte {
	conn, err := net.Dial("tcp", conf.Domain+":80")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Send HTTP GET request
	request := fmt.Sprintf("POST %s HTTP/1.1\r\n", conf.Endpoints[endpointKey])
	request += fmt.Sprintf("Host: %s\r\n", conf.Domain)
	request += "Content-Length: " + strconv.Itoa(len(body)) + "\r\n"
	request += "Connection: close\r\n"
	request += "\r\n"
	request += string(body) // Add the JSON body

    //utils.Print(request)

	_, err = conn.Write([]byte(request))

    resp, err := io.ReadAll(conn) // Read everything until connection closes
	if err != nil {
		panic(err)
	}

    //utils.Print(string(resp))


	return resp
}

func ParseResp(resp[] byte) (status string, body []byte, err error) {
    parts := bytes.SplitN(resp, []byte("\r\n\r\n"), 2)
    if len(parts) < 2 {
        return "", nil, errors.New("Unable to Split 1")
    }

    body = parts[1]

    headers := parts[0]

    if checkIfChunked(headers) {
        body, _ = fixIfChunked(body)
        utils.Print("Request is chuncked")
        utils.Print("Chuncked Body -> ", string(body))
    }

    statusLine := bytes.SplitN(headers, []byte("\r\n"), 2)
    if len(parts) < 2 {
        return "", nil, errors.New("Unable to Split 2 ")
    }

    statusLineComp := bytes.SplitN(statusLine[0], []byte(" "), 3)
    if len(statusLineComp) < 3 {
        return "", nil, errors.New("Unable to Split 3")
    }
    status = string(statusLineComp[1])

    return status, body, nil

}

func checkIfChunked(headers []byte) bool {
	return bytes.Contains(bytes.ToLower(headers), []byte("transfer-encoding: chunked"))
}

func fixIfChunked(body []byte) (realBody []byte, err error) {
	var result []byte
	reader := bytes.NewReader(body)

	for {
		// Read chunk size (hex value until \r\n)
		var chunkSizeHex []byte
		for {
			b := make([]byte, 1)
			_, err := reader.Read(b)
			if err != nil {
				return nil, err
			}
			if b[0] == '\r' {
				reader.Read(b) // Read '\n'
				break
			}
			chunkSizeHex = append(chunkSizeHex, b[0])
		}

		// Convert hex size to integer
		chunkSize, err := strconv.ParseInt(string(chunkSizeHex), 16, 64)
		if err != nil {
			return nil, err
		}

		// End of chunks (size 0)
		if chunkSize == 0 {
			break
		}

		// Read chunk data
		chunk := make([]byte, chunkSize)
		_, err = reader.Read(chunk)
		if err != nil {
			return nil, err
		}

		// Append to result
		result = append(result, chunk...)

		// Read \r\n after chunk data
		reader.Read(make([]byte, 2))
	}

	return result, nil
}
