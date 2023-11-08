package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"net/http"
)

const PORT int16 = 8080
const BAD_STR string = "monitoramento"

func handleConnection(conn net.Conn, sysLog *syslog.Writer) {
	remoteAddr := conn.RemoteAddr().String()
	msg := fmt.Sprintf("Connection received from: %s", remoteAddr)
	sysLog.Info(msg)
	fmt.Println(msg)

	defer conn.Close()

	buffer := make([]byte, 1024)

	n, err := conn.Read(buffer)
	if err != nil {
		msg := fmt.Sprintf("Failed to read data. %s", err)
		sysLog.Err(msg)
		fmt.Println(msg)
		return
	}
	fmt.Printf("Received: %s\n", buffer[:n])

	if bytes.Contains(buffer[:n], []byte(BAD_STR)) {
		msg := fmt.Sprintf("Unauthorized access for ip %s", remoteAddr)
		fmt.Printf(msg)
		sysLog.Info(msg)

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
		conn.Write([]byte(res))
	} else {
		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buffer[:n])))
		if err != nil {
			msg := fmt.Sprintf("Failed to parse HTTP request: %s", err)
			sysLog.Err(msg)
			fmt.Println(msg)
			return
		}

		targetServer := req.Host

		// Forward the data to the server
		serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:80", targetServer))
		if err != nil {
			sysLog.Err(fmt.Sprintf("Failed to connect to server: %s", err))
			fmt.Println(err)
			return
		}

		defer serverConn.Close()

		// Write data to the server
		_, err = serverConn.Write(buffer[:n])
		if err != nil {
			sysLog.Err(fmt.Sprintf("Failed to write to server: %s", err))
			fmt.Println(err)
			return
		}

		// Read the server's response
		serverBuffer := make([]byte, 4096)
		serverDataLen, err := serverConn.Read(serverBuffer)
		if err != nil {
			sysLog.Err(fmt.Sprintf("Failed to read from server: %s", err))
			fmt.Println(err)
			return
		}

		// Forward the server's response back to the client
		_, err = conn.Write(serverBuffer[:serverDataLen])
		if err != nil {
			sysLog.Err(fmt.Sprintf("Failed to write back to client: %s", err))
			fmt.Println(err)
			return
		}
	}
}

func main() {
	sysLog, err := syslog.New(syslog.LOG_LOCAL7|syslog.LOG_DEBUG, "Proxy")
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
