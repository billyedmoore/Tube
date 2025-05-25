package server

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"sync"

	"github.com/billyedmoore/tube/internal/websocket"
)

type share struct {
	shareCode          [5]byte
	senderConnection   *websocket.Connection
	receiverConnection *websocket.Connection
}

type globalContext struct {
	lock                    sync.Mutex
	activeShares            []*share
	sharesAwaitingReceivers map[[5]byte]*share
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
		panic("Failed to create connection.")
	}

	err = websocket.UpgradeConnection(w, r, connection)

	if err != nil {
		panic("Failed to upgrade connection.")
	}

	newShare, err := createShare(connection)

	if err != nil {
		panic("Failed to create share.")
	}

	h.context.lock.Lock()
	h.context.sharesAwaitingReceivers[newShare.shareCode] = newShare
	h.context.lock.Unlock()
}

func (h receiverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Have the share code in the HTTP request
	// Use the share code from the HTTP request to find the share
	// Upgrade the connection at share.receiverConnection
	// Move the share from shares awaiting receivers to activeShares
}

func createShare(senderConnection *websocket.Connection) (*share, error) {
	receiverConnection, err := websocket.CreateConnection()

	if err != nil {
		return nil, fmt.Errorf("Failed to create receiver connection")
	}

	var shareCode [5]byte
	_, err = rand.Read(shareCode[:])
	// TODO: check is not already taken however unlikely

	if err != nil {
		return nil, fmt.Errorf("Random bytes failed")
	}

	newShare := &share{
		shareCode:          shareCode,
		senderConnection:   senderConnection,
		receiverConnection: receiverConnection,
	}

	// here we should spin up a go routine to handle the actual sharing
	return newShare, nil
}
