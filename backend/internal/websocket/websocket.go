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
	"time"
)

type Connection struct {
	Incoming                      chan []byte
	lock                          sync.Mutex
	connected                     bool
	connectionStatusChangedSignal *sync.Cond
	conn                          net.Conn
	closing                       bool
	closeRetryTime                time.Duration
	closeGiveUpTime               time.Duration
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
			if (err == nil) && (n != 0) {
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
					sendPongFrame(connection, frm)
				case CLOSE_FRAME:
					fmt.Println("CLOSE_FRAME")

					if IsClosing(connection) {
						closeServer(connection)
					} else {
						sendCloseFrame(connection)
						closeServer(connection)
					}

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
		incoming:        make(chan []byte, 64),
		connected:       false,
		closeRetryTime:  time.Second * 2,
		closeGiveUpTime: time.Second * 30,
	}

	connection.connectionStatusChangedSignal = sync.NewCond(&connection.lock)

	return connection, nil
}

// Update the connection with the underlying connection
func instantiateConnection(connection *Connection, conn net.Conn) error {
	connection.lock.Lock()
	defer connection.lock.Unlock()
	connection.conn = conn
	connection.connected = true
	connection.connectionStatusChangedSignal.Signal()

	// Handle incoming messages as they come
	go readWorker(connection)

	return nil
}

func write(connection *Connection, data []byte) error {
	// TODO: look into the implications of partial writes
	connection.lock.Lock()
	defer connection.lock.Unlock()
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

// Doesn't send any Close Frames should be used after close handshake is
// complete
func closeServer(connection *Connection) error {
	// Closing the channel will kill the workers
	close(connection.incoming)
	connection.lock.Lock()
	defer connection.lock.Unlock()
	connection.connected = false
	connection.connectionStatusChangedSignal.Signal()
	return connection.conn.Close()
}

func decodeFrame(recievedData []byte) (frame, error) {
	if len(recievedData) < 2 {
		return frame{}, fmt.Errorf("Not a valid frame, not enough bytes.")
	}

	var fin bool = (recievedData[0] & 0x80) != 0
	var operation opcode = opcode((recievedData[0] & 0x0F))
	var mask bool = (recievedData[1] & 0x80) != 0
	var payloadLength uint64 = uint64((recievedData[1]) & 0x7F)
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
		err := fmt.Errorf(
			"Not a valid frame, not enough bytes (payload length %v shorter than specified payload length %v).",
			uint64(len(recievedData[nextByteIndex:])), payloadLength)
		return frame{}, err
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

func SendBlobData(connection *Connection, data []byte) error {
	if !IsConnected(connection) {
		return fmt.Errorf("Connection not connected.")
	}
	if IsClosing(connection) {
		return fmt.Errorf("Connection is closing.")
	}

	frm, err := newBinaryFrame(data)

	if err != nil {
		return fmt.Errorf("Couldn't create binary frame for data: %v.", data)
	}

	payload, err := encodeFrame(frm)

	if err != nil {
		return fmt.Errorf("Couldn't encode binary frame for data: %v.", data)
	}

	err = write(connection, payload)

	if err != nil {
		return fmt.Errorf("Couldn't write binary frame for data: %v.", data)
	}

	return nil
}

// This is the external class to allow the inititation of a close by external users
// TODO: design such that if there are errors sending the close frame there is visibility
func InitiateClose(connection *Connection) error {
	fmt.Println("CLOSE INITATED")

	if !IsConnected(connection) {
		return fmt.Errorf("Connection not connected.")
	}
	if IsClosing(connection) {
		return fmt.Errorf("Close already initiated.")
	}

	// Reattempts after "retryTime" and gives up and closes anyway after a further wait of
	// "giveUpTime"
	backgroundCloseFrameRetry := func() {
		connection.lock.Lock()
		retryTime := connection.closeRetryTime
		giveUpTime := connection.closeGiveUpTime
		connection.lock.Unlock()

		time.Sleep(retryTime)
		if IsConnected(connection) {
			sendCloseFrame(connection)
			// If after waiting give up time we haven't recieved a CloseFrame from the client close anyway
			time.Sleep(giveUpTime)
			if IsConnected(connection) {
				closeServer(connection)
			}
		}

	}

	err := sendCloseFrame(connection)
	if err != nil {
		return err
	}

	go backgroundCloseFrameRetry()
	return nil
}

func sendCloseFrame(connection *Connection) error {
	fmt.Println("SENDING CLOSE FRAME")
	if !IsConnected(connection) {
		return fmt.Errorf("Connection not connected.")
	}

	frm, err := newCloseFrame()

	if err != nil {
		return fmt.Errorf("Couldn't create close frame.")
	}

	payload, err := encodeFrame(frm)

	if err != nil {
		return fmt.Errorf("Couldn't encode close frame %v.", frm)
	}

	err = write(connection, payload)

	if err != nil {
		return fmt.Errorf("Couldn't write frame %v.", frm)
	}

	connection.lock.Lock()
	defer connection.lock.Unlock()
	connection.closing = true

	return nil
}

func sendPongFrame(connection *Connection, pingFrame frame) error {
	if !IsConnected(connection) {
		return fmt.Errorf("Connection not connected.")
	}

	pong, err := newPongFrame(pingFrame.payload)
	if err != nil {
		return fmt.Errorf("Failed to create pong frame.")
	}
	pongFrameBytes, err := encodeFrame(pong)
	if err != nil {
		return fmt.Errorf("Failed to encode pong frame %v.", pong)
	}
	err = write(connection, pongFrameBytes)
	if err != nil {
		fmt.Println("Failed to write pong frame.")
		return fmt.Errorf("Failed to write ping frame %v encoded as %v.", pong, pongFrameBytes)
	}
	return nil
}

func sendPingFrame(connection *Connection, pingFrame frame) error {
	if !IsConnected(connection) {
		return fmt.Errorf("Connection not connected.")
	}

	if IsClosing(connection) {
		return fmt.Errorf("Connection closing, can't ping.")
	}

	ping, err := newPingFrame()
	if err != nil {
		return fmt.Errorf("Failed to create ping frame.")
	}
	pingFrameBytes, err := encodeFrame(ping)
	if err != nil {
		return fmt.Errorf("Failed to encode ping frame %v.", ping)
	}
	err = write(connection, pingFrameBytes)
	if err != nil {
		return fmt.Errorf("Failed to write ping frame %v encoded as %v.", ping, pingFrameBytes)
	}
	return nil
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

	// TODO: different error codes for different senarios
	binary.BigEndian.PutUint16(code, 1000)
	buffer.Write(code)
	buffer.Write([]byte("Connection closed normally."))
	payload := buffer.Bytes()

	frm := frame{fin: true, operation: CLOSE_FRAME,
		mask: true, payloadLength: uint64(len(payload)), payload: payload}

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

// Do nothing just wait until the connection is connected
func WaitUntilConnected(connection *Connection) {
	connection.lock.Lock()
	defer connection.lock.Unlock()
	// wait for the connection to be connected
	for !connection.connected {
		connection.connectionStatusChangedSignal.Wait()
	}
}

// Is the connection obj connected and ready to send data
func IsConnected(connection *Connection) bool {
	connection.lock.Lock()
	defer connection.lock.Unlock()
	return connection.connected
}

// Is the connection obj in the process of closing
func IsClosing(connection *Connection) bool {
	connection.lock.Lock()
	defer connection.lock.Unlock()
	return connection.closing
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
