package customhttp

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"orsted/beacon/utils"
	"strings"
)

// shared helper to build the CONNECT request
func buildConnectRequest(targetAddr, username, password string) string {
	authHeader := ""
	if username != "" && password != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		authHeader = fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", auth)
	}
	return fmt.Sprintf(
		"CONNECT %s HTTP/1.1\r\nHost: %s\r\n%s\r\n",
		targetAddr, targetAddr, authHeader,
	)
}

// shared helper to read and validate proxy response
func validateProxyResponse(conn net.Conn) error {
	reader := bufio.NewReader(conn)

	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.Contains(statusLine, "200") {
		return fmt.Errorf("proxy CONNECT failed: %s", strings.TrimSpace(statusLine))
	}

	// consume headers until blank line
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if line == "\r\n" {
			break
		}
	}
	return nil
}

// DialHTTPProxy dials a plain HTTP proxy (no TLS on proxy itself)
func DialHTTPProxy(proxyAddr, targetAddr, username, password string) (net.Conn, error) {
	utils.Print("Calling ", targetAddr, "via proxy http ", proxyAddr)
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, err
	}

	utils.Print("Done.")
	req := buildConnectRequest(targetAddr, username, password)
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}
	utils.Print("Done.")

	utils.Print("Validating Proxy Response to above req")

	if err := validateProxyResponse(conn); err != nil {
		conn.Close()
		return nil, err
	}
	utils.Print("Done.")

	return conn, nil
}

// DialHTTPSProxy dials an HTTPS proxy (TLS connection to proxy)
func DialHTTPSProxy(proxyAddr, targetAddr, username, password string) (net.Conn, error) {
	utils.Print("Calling ", targetAddr, "via proxy https ", proxyAddr)
	conn, err := tls.Dial("tcp", proxyAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}

	utils.Print("Done.")
	req := buildConnectRequest(targetAddr, username, password)
	utils.Print("Building and writing CONNECT to proxy with following req", req)
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}
	utils.Print("Done.")

	utils.Print("Validating Proxy Response to above req")
	if err := validateProxyResponse(conn); err != nil {
		conn.Close()
		return nil, err
	}
	utils.Print("Done.")

	return conn, nil
}

