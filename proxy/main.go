package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net"
	"net/http"
)

const (
	PORT    int16  = 8080
	host    string = "127.0.0.1"
	BAD_STR string = "monitorando"
)

func blockAccess(remoteAddr string) []byte {

	res := "HTTP/1.1 200 OK\n" +
		"Server: Microsoft-IIS/4.0" +
		"Date: Mon, 3 Jan 2016 17:13:34 GMT\n" +
		"Content-Type: text/html; charset=utf-8\n" +
		"Last-Modified: Mon, 11 Jan 2016 17:24:42 GMT\n" +
		"Content-Length: 112\n\n" +
		"<html>\n" +
		"<head>\n" +
		"<title>Exemplo de resposta HTTP</title>\n" +
		"</head>\n" +
		"<body>Acesso não autorizado!</body>\n" +
		"</html>\n\n"

	return []byte(res)
}

func getHost(buffer []byte) (string, error) {
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buffer)))

	if err != nil {
		return "", err
	}

	targetServer := req.URL.Hostname()
	return fmt.Sprintf("%s:80", targetServer), nil
}

func connectToServer(host string) (net.Conn, error) {
	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	return serverConn, nil
}

func proxy(buffer []byte, clientConn net.Conn, host string) (int, error) {
	serverConn, err := connectToServer(host)
	if err != nil {
		return 0, err
	}
	defer serverConn.Close()

	n, err := serverConn.Write(buffer)
	if err != nil {
		return n, err
	}

	serverBuffer := make([]byte, 4096)
	for {
		n, err = serverConn.Read(serverBuffer)
		if err != nil {
			return n, err
		}

		n, err = clientConn.Write(serverBuffer)
		if err != nil {
			return n, err
		}
	}
}

func handleConnection(conn net.Conn, sysLog *syslog.Writer) {
	remoteAddr := conn.RemoteAddr().String()
	msg := fmt.Sprintf("Connection received from: %s", remoteAddr)
	sysLog.Info(msg)
	fmt.Println(msg)

	defer conn.Close()

	for {
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				msg := fmt.Sprintf("Client disconnected: %s", remoteAddr)
				sysLog.Info(msg)
				fmt.Println(msg)
				break
			}

			msg := fmt.Sprintf("Failed to read data: %s, %s", remoteAddr, err)
			sysLog.Err(msg)
			fmt.Println(msg)
			break
		}

		host, err := getHost(buffer[:n])
		if err != nil {
			msg := fmt.Sprintf("Failed to extract host from request: %s, %s", remoteAddr, err)
			sysLog.Err(msg)
			fmt.Println(msg)
			break
		}

		if bytes.Contains(buffer[:n], []byte(BAD_STR)) {
			msg := fmt.Sprintf("Unauthorized access from IP: %s to %s", remoteAddr, host)
			fmt.Println(msg)
			sysLog.Info(msg)
			conn.Write(blockAccess(remoteAddr))
			break
		} else {
			go func() {
				msg := fmt.Sprintf("Proxying from %s to %s", remoteAddr, host)
				sysLog.Info(msg)
				fmt.Println(msg)

				_, err = proxy(buffer[:n], conn, host)
				if err != nil {
					if err == io.EOF {
						msg := fmt.Sprintf("Server disconnected: %s", host)
						sysLog.Info(msg)
						fmt.Println(msg)
					} else {
						sysLog.Err(fmt.Sprintf("Failed to proxy the request: %s", err))
						fmt.Println(err)
					}
				}
			}()
		}
	}

	msg = fmt.Sprintf("Closing connection with: %s", remoteAddr)
	sysLog.Info(msg)
	fmt.Println(msg)
}

func main() {
	sysLog, err := syslog.Dial("udp", "syslogserver:514", syslog.LOG_INFO|syslog.LOG_DAEMON, "proxy")
	if err != nil {
		log.Fatal("Error connecting to syslog")
	}

	defer sysLog.Close()
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", PORT))
	if err != nil {
		msg := fmt.Sprintf("Failed to listen on port %d. %s", PORT, err)
		sysLog.Err(msg)
		log.Fatal(msg)
	}
	defer ln.Close()

	msg := fmt.Sprintf("Listening on port %d", PORT)
	fmt.Println(msg)
	sysLog.Info(msg)

	for {
		conn, err := ln.Accept()
		if err != nil {
			msg := fmt.Sprintf("Failed to accept connection. %s", err)
			sysLog.Err(msg)
			fmt.Println(msg)
		}

		go handleConnection(conn, sysLog)
	}
}
