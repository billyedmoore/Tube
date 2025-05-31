package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"github.com/billyedmoore/tube/internal/websocket"
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
		return false, "share_code parameter is not set or is set to \"\"."
	}
	if len(shareCode) > 8 {
		return false, "Provided share_code is too long to be a valid share code."
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
		http.Error(w, "Provided share_code could not be decoded.", http.StatusBadRequest)
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

	go facilitateShare(newShare)

	return newShare, nil
}

func facilitateShare(share *Share) {
	senderInitiation := <-share.senderConnection.Incoming
	// TODO: Decode senderInitiation
	// TODO: Encode and send senderAcceptance to sender
	recieverInitiation := <-share.receiverConnection.Incoming
	// TODO: Decode recieverInitiation
	// TODO: Encode and send recieverAcceptance

	// TODO: Encode and send Ready to sender
	metaData := <-share.senderConnection.Incoming
	// TODO: Decode MetaData (To get number of chunks)
	// TODO: Forward MetaData

	// for i: = 0; i<n; i++{
	// TODO: Forward Chunk
	// TODO: Forward Chunk Ack
	//}
	//TODO: Close and update the context

}
