package websocket

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

type Connection struct {
	lock       sync.Mutex
	incoming   chan []byte
	connected  bool
	statusCond *sync.Cond
	conn       net.Conn
}

type frame struct {
	fin           bool
	operation     opcode
	mask          bool
	maskKey       uint32
	payloadLength uint64
	// event if frame is masked payload should be unmasked
	payload []byte
}

type opcode uint8

const (
	CONTINUATION_FRAME opcode = 0x0
	TEXT_FRAME         opcode = 0x1
	BINARY_FRAME       opcode = 0x2
	CLOSE_FRAME        opcode = 0x8
	PING_FRAME         opcode = 0x9
	PONG_FRAME         opcode = 0xA
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
	fmt.Println("\"" + challengeString + "\"")
	// Specified in RFC 6455

	str := challengeString + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	data := sha1.Sum([]byte(str))
	return base64.StdEncoding.EncodeToString(data[:])
}

func readWorker(connection *Connection) {
	buffer := make([]byte, 4096)
	for {
		select {
		case _, ok := <-connection.incoming:
			if !ok {
				fmt.Println("Channel is closed stopping ReadWorker")
				return
			}
		default:
			n, err := connection.conn.Read(buffer)
			if (err != nil) && (n != 0) {
				data := make([]byte, n)
				copy(data, buffer[:n])
				frm, err := decodeFrame(data)

				if err != nil {
					fmt.Println(hex.EncodeToString(data))
					fmt.Printf("Failed to decode frame. %e", err)
				}

				fmt.Printf("Frame recieved %v", frm.operation)
				switch frm.operation {
				case BINARY_FRAME:
					connection.incoming <- frm.payload
				case PING_FRAME:
					pong, err := newPongFrame(frm.payload)
					if err != nil {
						fmt.Println("Failed to create pong frame.")
					}
					pongFrameBytes, err := encodeFrame(pong)
					if err != nil {
						fmt.Println("Failed to encode pong frame.")
					}
					err = Write(connection, pongFrameBytes)
					if err != nil {
						fmt.Println("Failed to write pong frame.")
					}
				case CLOSE_FRAME:
					// if we recieve a close send one back
					// TODO: make sure we are returning the correct status code
					fmt.Println("CLOSE_FRAME")
					Close(connection)

				default:
					fmt.Println("Ignored recieved data. ")
				}
			}
		}
	}
}

// New connection object or error (no errors yet)
func CreateConnection() (*Connection, error) {
	connection := &Connection{
		incoming:  make(chan []byte),
		connected: false,
	}

	connection.statusCond = sync.NewCond(&connection.lock)

	return connection, nil
}

// Update the connection with the underlying connection
func instantiateConnection(connection *Connection, conn net.Conn) error {
	connection.lock.Lock()
	connection.conn = conn
	connection.connected = true
	connection.statusCond.Signal()
	connection.lock.Unlock()

	// Handle incoming messages as they come
	go readWorker(connection)

	return nil
}

func Write(connection *Connection, data []byte) error {
	// TODO: look into the implications of partial writes

	fmt.Printf("Writing %v bytes to connection data : %v\n", len(data), data)
	writtenBytes := 0
	for writtenBytes < len(data) {
		n, err := connection.conn.Write(data[writtenBytes:])
		if err != nil {
			return err
		}
		writtenBytes += n
	}
	return nil

}

func Close(connection *Connection) error {
	frm, err := newCloseFrame()
	if err != nil {
		return err
	}
	data, err := encodeFrame(frm)
	if err != nil {
		return err
	}
	err = Write(connection, data)
	if err != nil {
		return err
	}
	// Closing the channel will kill the workers
	close(connection.incoming)
	connection.lock.Lock()
	connection.connected = false
	connection.statusCond.Signal()
	connection.lock.Unlock()
	err = connection.conn.Close()
	return err
}

func decodeFrame(recievedData []byte) (frame, error) {
	if len(recievedData) < 2 {
		return frame{}, fmt.Errorf("Not a valid frame, not enough bytes.")
	}

	var fin bool = (recievedData[0] & 1) != 0
	var operation opcode = opcode((recievedData[0] << 4) & 0xF0)
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
	var payload []byte = make([]byte, payloadLength)
	if mask {
		payloadBytes, err := applyMask(maskKey, recievedData[nextByteIndex:])
		if err != nil {
			return frame{}, fmt.Errorf("Masking failed.")
		}
		copy(payload, payloadBytes)
	} else {
		copy(payload, recievedData[nextByteIndex:])
	}

	data := frame{fin: fin, operation: operation, mask: mask,
		maskKey: maskKey, payloadLength: payloadLength, payload: payload}
	return data, nil
}

func newBinaryFrame(data []byte) (frame, error) {
	payload := make([]byte, len(data))
	copy(payload, data)
	frm := frame{fin: true, operation: BINARY_FRAME,
		mask: false, payloadLength: uint64(len(data)), payload: payload}

	return frm, nil
}

func newPongFrame(data []byte) (frame, error) {
	payload := make([]byte, len(data))
	copy(payload, data)
	frm := frame{fin: true, operation: PONG_FRAME,
		mask: false, payloadLength: uint64(len(data)), payload: payload}

	return frm, nil
}

func newPingFrame() (frame, error) {
	payload := []byte("Ping!") // optional content
	frm := frame{fin: true, operation: PING_FRAME,
		mask: false, payloadLength: uint64(len(payload)), payload: payload}

	return frm, nil
}

func newCloseFrame() (frame, error) {
	var buffer bytes.Buffer
	code := make([]byte, 2)

	// normal closure
	// TODO: select error code situationally
	binary.BigEndian.PutUint16(code, 1000)
	buffer.Write([]byte("Connection closed normally."))
	payload := buffer.Bytes()

	frm := frame{fin: true, operation: CLOSE_FRAME,
		mask: false, payloadLength: uint64(len(payload)), payload: payload}

	return frm, nil
}

func applyMask(maskKey uint32, payload []byte) ([]byte, error) {
	mask := make([]byte, 4)
	binary.BigEndian.PutUint32(mask, maskKey)

	maskedPayload := make([]byte, len(payload))

	for index, value := range payload {
		maskedPayload[index] = value ^ mask[index%4]
	}

	return maskedPayload, nil
}

func encodeFrame(data frame) ([]byte, error) {
	var fin uint8 = 0

	if data.fin {
		fin = 1
	}

	var mask uint8 = 0

	if data.mask {
		mask = 1
	}

	var op uint8 = uint8(data.operation)
	var firstByte uint8 = op + (fin << 7)

	var payloadLengthBytes []byte
	var payloadLength7bit uint8

	if data.payloadLength <= 125 {
		payloadLength7bit = uint8(data.payloadLength)
		payloadLengthBytes = make([]byte, 0)
	} else if data.payloadLength < 65535 {
		payloadLength7bit = uint8(126)
		payloadLengthBytes = make([]byte, 2)
		binary.BigEndian.PutUint16(payloadLengthBytes, uint16(data.payloadLength))
	} else {
		payloadLength7bit = uint8(127)
		payloadLengthBytes = make([]byte, 8)
		binary.BigEndian.PutUint64(payloadLengthBytes, uint64(data.payloadLength))
	}

	var secondByte uint8 = (mask << 7) + payloadLength7bit
	var maskKeyBytes []byte

	if data.mask {
		maskKeyBytes = make([]byte, 4)
		binary.BigEndian.PutUint32(maskKeyBytes, data.maskKey)
	} else {
		maskKeyBytes = make([]byte, 0)
	}
	var payloadBytes []byte

	if !data.mask {
		payloadBytes = make([]byte, data.payloadLength)
		copy(payloadBytes, data.payload)
	} else {
		var err error
		payloadBytes, err = applyMask(data.maskKey, data.payload)
		if err != nil {
			return nil, err
		}
	}

	var buffer bytes.Buffer

	buffer.Write([]byte{firstByte, secondByte})
	buffer.Write(payloadLengthBytes)
	buffer.Write(maskKeyBytes)
	buffer.Write(payloadBytes)

	println("encodedFrame :", buffer.Bytes())
	return buffer.Bytes(), nil

}

// Upgrade from http -> websocket, hijacks the connection if successful
// We dont
func UpgradeConnection(w http.ResponseWriter, r *http.Request, connection *Connection) error {
	println("Upgrading the connection")

	var challengeKey string = r.Header.Get("Sec-Websocket-Key")

	if (!checkHeader(r, "Upgrade", "websocket")) ||
		(!checkHeader(r, "Connection", "Upgrade")) ||
		(!checkHeader(r, "Sec-Websocket-Version", "13")) {
		return fmt.Errorf("Invalid headers for websocket upgrade.")
	}

	if !isValidChallengeKey(challengeKey) {
		return fmt.Errorf("Invalid challenge key.")
	}

	if !(r.Method == http.MethodGet) {
		return fmt.Errorf("Can only upgrade GET requests.")
	}

	wAsHijacker, ok := w.(http.Hijacker)

	if !ok {
		return fmt.Errorf("Failed to Hijack.")
	}

	underlyingConnection, buffer, err := wAsHijacker.Hijack()

	if err != nil {
		return err
	}

	instantiateConnection(connection, underlyingConnection)

	// Build the response
	response := []string{
		"HTTP/1.1 101 Switching Protocols",
		"Upgrade: websocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Accept: " + generateAcceptKey(challengeKey),
		"",
		"",
	}

	fmt.Println(strings.Join(response, "\r\n"))
	_, err = buffer.WriteString(strings.Join(response, "\r\n"))

	//TODO: Consider if there is a way of handling these such that the client
	//	can still be sent an error message, current behaviour is the connection
	//	is hijacked so sending anything to w is unusable.

	if err != nil {
		return err
	}

	err = buffer.Flush()

	if err != nil {
		return err
	}
	println("Upgraded the connection")
	return nil
}
