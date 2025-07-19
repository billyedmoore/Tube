package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"github.com/billyedmoore/tube/internal/websocket"
)

type opcode uint8

const highestOpCode = 0x9

const (
	SENDER_INITIATION   opcode = 0x1
	SENDER_ACCEPTED     opcode = 0x2
	RECEIVER_INITIATION opcode = 0x3
	RECIEVER_ACCEPTED   opcode = 0x4
	READY               opcode = 0x5
	METADATA            opcode = 0x6
	DATA_CHUNK          opcode = 0x7
	AWKNOWLEDGE         opcode = 0x8
	ERROR               opcode = 0x9
)

type Share struct {
	shareCode          [5]byte
	senderConnection   *websocket.Connection
	receiverConnection *websocket.Connection
}

type globalContext struct {
	lock                    sync.Mutex
	activeShares            map[[5]byte]*Share
	sharesAwaitingReceivers map[[5]byte]*Share
}

type senderHandler struct {
	context *globalContext
}

type receiverHandler struct {
	context *globalContext
}

func (h senderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	connection, err := websocket.CreateConnection()

	if err != nil {
		http.Error(w, "Interal server error.", http.StatusInternalServerError)
		return
	}

	err = websocket.UpgradeConnection(w, r, connection)

	if err != nil {
		http.Error(w, "Websocket failed to upgrade.", http.StatusInternalServerError)
		return
	}

	newShare, err := createShare(connection, h.context)

	if err != nil {
		//TODO: send error frame over websocket
		websocket.InitiateClose(connection)
		return
	}

	h.context.lock.Lock()
	h.context.sharesAwaitingReceivers[newShare.shareCode] = newShare
	h.context.lock.Unlock()
}

func isValidShareCode(shareCode string) (bool, string) {
	if len(shareCode) == 0 {
		return false, "shareCode parameter is not set or is set to \"\"."
	}
	if len(shareCode) > 8 {
		return false, "Provided shareCode is too long to be a valid share code."
	}
	return true, ""
}

func (h receiverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	encodedShareCode := r.URL.Query().Get("share_code")

	isValid, reason := isValidShareCode(encodedShareCode)

	if !isValid {
		http.Error(w, reason, http.StatusBadRequest)
		return
	}

	shareCodeSlice, err := base64.StdEncoding.DecodeString(encodedShareCode)

	if err != nil {
		http.Error(w, "Provided shareCode could not be decoded.", http.StatusBadRequest)
		return
	}

	var shareCode [5]byte
	copy(shareCode[:], shareCodeSlice)

	h.context.lock.Lock()
	defer h.context.lock.Unlock()
	share := h.context.sharesAwaitingReceivers[shareCode]
	delete(h.context.sharesAwaitingReceivers, shareCode)
	h.context.activeShares[shareCode] = share
	h.context.lock.Unlock()

	err = websocket.UpgradeConnection(w, r, share.receiverConnection)

	if err != nil {
		http.Error(w, "Websocket failed to upgrade.", http.StatusInternalServerError)
		return
	}

}

func createShare(senderConnection *websocket.Connection, context *globalContext) (*Share, error) {
	receiverConnection, err := websocket.CreateConnection()

	if err != nil {
		return nil, fmt.Errorf("Failed to create receiver connection")
	}

	var shareCode [5]byte
	shareCodeSet := false

	context.lock.Lock()
	defer context.lock.Unlock()

	for !shareCodeSet {
		_, err = rand.Read(shareCode[:])

		if err != nil {
			return nil, fmt.Errorf("Random bytes failed")
		}
		_, shareCodeUsedByActiveShare := context.activeShares[shareCode]
		_, shareCodeUsedByNewShare := context.sharesAwaitingReceivers[shareCode]

		if (!shareCodeUsedByActiveShare) && (!shareCodeUsedByNewShare) {
			shareCodeSet = true
		}
	}

	newShare := &Share{
		shareCode:          shareCode,
		senderConnection:   senderConnection,
		receiverConnection: receiverConnection,
	}

	// start the go-routine that will handle the share
	go facilitateShare(newShare, context)

	return newShare, nil
}

func errorOutShare(share *Share, context *globalContext, errorReason string) {
	const maxLength = 65535

	if len(errorReason) > maxLength {
		errorReason = errorReason[:maxLength]
	}

	errorEncoded, err := encodeError(errorReason)

	if err != nil {
		// encodeError only returns an error for input too long
		// since this case has been handled this should never happen
		panic("ErrorReason should not be too long but is.")
	}

	if websocket.IsConnected(share.senderConnection) {
		err = websocket.SendBlobData(share.senderConnection, errorEncoded)
		if err != nil {
			// failed to send error to sender
		}
	}
	if websocket.IsConnected(share.receiverConnection) {
		err = websocket.SendBlobData(share.senderConnection, errorEncoded)
		if err != nil {
			// failed to send error to reciever
		}
	}
	websocket.InitiateClose(share.senderConnection)
	websocket.InitiateClose(share.receiverConnection)

	context.lock.Lock()
	defer context.lock.Unlock()
	// delete the last reference to the share
	// effectively this is the free() point
	delete(context.activeShares, share.shareCode)
}

func facilitateShare(share *Share, context *globalContext) {
	// TODO: refactor this into smaller functions
	// TODO: implement the many Unimplemented edgecases

	websocket.WaitUntilConnected(share.senderConnection)
	senderInitiation := <-share.senderConnection.Incoming
	err := decodeSenderInitiation(senderInitiation)

	if err != nil {
		panic("Unimplmented edgecase")
		// close the share and clean up
	}

	senderAcceptance, err := encodeSenderAcceptance(share.shareCode[:])

	if err != nil {
		panic("Unimplmented edgecase")
		// close the share and clean up
	}

	err = websocket.SendBlobData(share.senderConnection, senderAcceptance)

	if err != nil {
		// close the share
		// communicate with client
		// clean up
		panic("Unimplmented edgecase")
	}

	websocket.WaitUntilConnected(share.receiverConnection)
	recieverInitiation := <-share.receiverConnection.Incoming

	recieverPublicKey, err := decodeReceiverInitiation(recieverInitiation)

	if err != nil {
		panic("Un-implmented edgecase")
		// close the share and clean up
	}

	recieverAcceptance := encodeRecieverAcceptance()
	err = websocket.SendBlobData(share.receiverConnection, recieverAcceptance)

	if err != nil {
		// close the share
		// communicate with sender
		// clean up
		panic("Unimplmented edgecase")
	}

	ready, err := encodeReady(recieverPublicKey)
	err = websocket.SendBlobData(share.senderConnection, ready)

	if err != nil {
		// close the share
		// communicate with sender
		// clean up
		panic("Unimplmented edgecase")
	}

	meta := <-share.senderConnection.Incoming
	numberOfChunks, err := decodeMetadata(meta)

	err = websocket.SendBlobData(share.receiverConnection, meta)

	if err != nil {
		// close the share
		// communicate with sender
		// clean up
		panic("Unimplmented edgecase")
	}

	metaDataAck := <-share.receiverConnection.Incoming
	chunkNumber, err := decodeAcknowledge(metaDataAck)

	if err != nil {
		panic("Un-implmented edgecase")
		// close the share and clean up
	}

	if chunkNumber != 0xFF {
		panic("Un-implmented edgecase")
		// close the share and clean up
	}

	err = websocket.SendBlobData(share.senderConnection, metaDataAck)

	if err != nil {
		// close the share
		// communicate with sender
		// clean up
		panic("Unimplmented edgecase")
	}

	for i := uint16(0); i <= numberOfChunks; i++ {
		chunk := <-share.senderConnection.Incoming

		chunkNumber, err = decodeDataChunk(chunk)

		if err != nil || chunkNumber != i {
			// Failed to decode data chunk or mismatched state bettween client and server
			// Probably send some sort of error to the clients
			panic("Un-implmented edgecase")
		}

		err = websocket.SendBlobData(share.receiverConnection, chunk)

		if err != nil {
			// close the share
			// communicate with sender
			// clean up
			panic("Unimplmented edgecase")
		}

		metaDataAck := <-share.receiverConnection.Incoming
		chunkNumber, err := decodeAcknowledge(metaDataAck)

		if err != nil || chunkNumber != i {
			// Failed to decode data chunk or mismatched state bettween client and server
			// Probably send some sort of error to the clients
			panic("Un-implmented edgecase")
		}

		err = websocket.SendBlobData(share.senderConnection, metaDataAck)

		if err != nil {
			// close the share
			// communicate with sender
			// clean up
			panic("Unimplmented edgecase")
		}

	}

	websocket.InitiateClose(share.senderConnection)
	websocket.InitiateClose(share.receiverConnection)

	context.lock.Lock()
	defer context.lock.Unlock()
	// delete the last reference to the share
	// effectively this is the free() point
	delete(context.activeShares, share.shareCode)
}
