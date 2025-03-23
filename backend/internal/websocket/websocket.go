package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
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
	incoming chan []byte

	conn net.Conn
}

func readWorker(connection *Connection) {
	buffer := make([]byte, 4096)
	for {
		select {
		case <-connection.incoming:
			fmt.Println("Channel is closed stopping ReadWorker")
			return
		default:
			n, err := connection.conn.Read(buffer)
			if err != nil {
				data := make([]byte, n)
				copy(data, buffer[:n])
				//TODO: parse out the actual data frame
				connection.incoming <- data
			}
		}
	}
}

// New connection object or error (no errors yet)
func newConnection(conn net.Conn) (*Connection, error) {
	connection := &Connection{
		incoming: make(chan []byte),
		conn:     conn}

	// handle Reads data as it comes in
	go readWorker(connection)
	return connection, nil
}

func Write(connection *Connection, data []byte) error {
	//TODO: implement
	return errors.New("Not implemented")
}

func Close(connection *Connection, data []byte) error {
	Write(connection, []byte{0x08})
	// Closing the channel will kill the workers
	close(connection.incoming)
	err := connection.conn.Close()
	return err
}

type frame struct {
	fin           bool
	opcode        uint8
	mask          bool
	maskKey       uint32
	payloadLength uint64
	payload       []byte
}

func parseFrame(recievedData []byte) (frame, error) {
	if len(recievedData) < 2 {
		return frame{}, fmt.Errorf("Not a valid frame, not enough bytes.")
	}

	var fin bool = (recievedData[0] & 1) != 0
	var opcode uint8 = (recievedData[0] << 4) & 0xF0
	var mask bool = (recievedData[1] & 1) != 0
	var payloadLength uint64 = uint64((recievedData[1] << 1) & 0xFE)
	var maskKey uint32 = 0

	nextByteIndex := 2

	switch payloadLength {
	case 126:
		payloadLength = binary.BigEndian.Uint64(recievedData[nextByteIndex:(nextByteIndex + 2)])
		nextByteIndex += 2
	case 127:
		payloadLength = binary.BigEndian.Uint64(recievedData[nextByteIndex:(nextByteIndex + 8)])
		nextByteIndex += 8
	}

	if mask {
		if len(recievedData[nextByteIndex:]) < 4 {
			return frame{}, fmt.Errorf("Not a valid frame, not enough bytes (mask true but no mask key).")
		}
		maskKey = binary.BigEndian.Uint32(recievedData[nextByteIndex:(nextByteIndex + 4)])
		nextByteIndex += 4
	}
	if uint64(len(recievedData[nextByteIndex:])) < payloadLength {
		return frame{}, fmt.Errorf("Not a valid frame, not enough bytes (payload shorter than payload length).")
	}
	var payload = make([]byte, payloadLength)
	copy(payload, recievedData[nextByteIndex:])

	data := frame{fin: fin, opcode: opcode, mask: mask,
		maskKey: maskKey, payloadLength: payloadLength, payload: payload}
	return data, nil
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
