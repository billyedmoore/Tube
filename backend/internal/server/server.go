package server

import (
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
	lock                   sync.Mutex
	activeShares           []*share
	sharesAwaitingRecivers map[[5]byte]*share
}

type senderHandler struct {
	context *globalContext
}

type receiverHandler struct {
	context *globalContext
}

func (h senderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (h receiverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}
