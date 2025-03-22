package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

func checkHeader(r *http.Request, key string, value string) bool {
	return (r.Header.Get(key) == value)
}

// Challenge key should be 16 character base64 encoded string
func isValidChallengeKey(str string) bool {

	if str == "" {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(str)

	if err != nil {
		return false
	} else {
		return (len(decoded) == 16)
	}
}

func generateAcceptKey(challengeString string) string {
	// Specified in RFC 6455
	guid := []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	data := sha1.Sum(append([]byte(challengeString), guid...))
	return base64.StdEncoding.EncodeToString(data[:])
}

func assert(conditional bool) {
	if !conditional {
		log.Fatalf("Assertion failed.")
	}
}

type Connection struct {
	// no lock atm as channels are thread safe

	// This is the interface for the connection.
	// Applications can expect any incoming data to be written to the toBeRead
	// channel and they can write bytes to toBeWritten channel and expect them
	// to be sent to the client.
	toBeRead    chan []byte
	toBeWritten chan []byte

	conn net.Conn
}

func readWorker(connection *Connection) {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-connection.toBeRead:
			fmt.Println("Channel is closed stopping ReadWorker")
			return
		default:
			n, err := connection.conn.Read(buffer)
			if err != nil {
				data := make([]byte, n)
				copy(data, buffer[:n])
				//TODO: parse out the actual data from the meta (ie bytes vs strings)
				connection.toBeRead <- data
			}
		}
	}
}

func writeWorker(connection *Connection) {
	for data := range connection.toBeWritten {
		Write(connection, data)
	}
}

// New connection object or error (no errors yet)
func newConnection(conn net.Conn) (*Connection, error) {
	connection := &Connection{
		toBeRead:    make(chan []byte),
		toBeWritten: make(chan []byte),
		conn:        conn}

	// handle Reads and Writes as they come in
	go readWorker(connection)
	go writeWorker(connection)
	return connection, nil
}

func Write(connection *Connection, data []byte) error {
	//TODO: implement
	return errors.New("Not implemented")
}

func Close(connection *Connection, data []byte) error {
	Write(connection, []byte{0x08})
	close(connection.toBeRead)
	close(connection.toBeWritten)
	err := connection.conn.Close()
	return err
}

// Upgrade from http -> websocket, hijacks the connection
func UpgradeConnection(w http.ResponseWriter, r *http.Request) *Connection {

	var challengeKey string = r.Header.Get("Sec-Websocket-Key")

	// Check the incoming request is a valid websocket upgrade and we can handle it
	//TODO: handle these errors gracefully, return an http error
	assert(checkHeader(r, "Upgrade", "websocket"))
	assert(checkHeader(r, "Connection", "Upgrade"))
	assert(checkHeader(r, "Sec-Websocket-Version", "13"))
	assert(isValidChallengeKey(challengeKey))
	assert(r.Method == http.MethodGet)

	wAsHijacker, ok := w.(http.Hijacker)
	assert(ok) // TODO: error handling

	underlyingConnection, buffer, err := wAsHijacker.Hijack()
	assert(err == nil) // TODO: error handling

	// Build the response
	response := []string{
		"HTTP/1.1 101 Switching Protocols",
		"Upgrade: websocket",
		"Connection : upgrade",
		"Sec-WebSocket-Accept: " + generateAcceptKey(challengeKey),
		"",
		"",
	}

	_, err = buffer.WriteString(strings.Join(response, "\r\n"))
	assert(err == nil) // TODO: error handling
	err = buffer.Flush()
	assert(err == nil) // TODO: error handling

	conn, err := newConnection(underlyingConnection)
	assert(err == nil) // TODO: error handling
	return conn
}
