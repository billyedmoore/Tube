# Backend

## Websockets

> [!WARNING]
> This is in early development. Information on this page may be out of date or factually incorrect.

Tube uses a custom websocket library for websocket communication.

Tube's websocket library is dependant only on go's standard library, networking details are handled by the "net" and "net/http" packages.

### Functions

+ `CreateConnection () -> *Connection, error`, create a new `Connection` object not yet linked to a connection.
+ `UpgradeConnection (http.ResponseWriter, *http.Request, *Connection) -> error`, hijack a HTTP connection and convert to websocket activating the `Connection` for usage.
+ `IsConnected (*Connection) -> bool`, is the `Connection` connected and ready to send and recieve data.
+ `IsClosing (*Connection) -> bool`, is the `Connection` in the process of closing, becomes `true` once a closing frame is sent.
+ `SendBlobData (*Connection, []byte) -> error`, send a blob of data over the websocket connection as a data frame.
+ `InitiateClose (*Connection) -> error`, send a close frame and set the state to closing so the server will close when it receives a close frame.
  Also starts a go routine that will resend the close frame after `connection.closeRetryTime` if one is not yet recieved from the client and attempt to close
  connection after `connection.closeGiveUpTime` if the connection is not yet closed.
+ `WaitUntilConnected (*Connection) -> nil`, waits until the connection is connected.

### The Connection Object

```go
type Connection struct {
	lock                          sync.Mutex
	incoming                      chan []byte
	connected                     bool
	connectionStatusChangedSignal *sync.Cond
	conn                          net.Conn
	closing                       bool
	closeRetryTime                time.Duration
	closeGiveUpTime               time.Duration
}
```

+ Where possible you should avoid accessing members of the Connection object directly instead using the
  above functions. If you do directly access or modify values `lock sync.Mutex` should be used.
+ `incoming chan []byte`, is the channel where all incoming data is written by the ReadWorker.

### Usage Basics

Hopfully usage is quite intuitive, a full example can be seen below:

```go

import ("github.com/billyedmoore/tube/websocket")

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	connection, err := CreateConnection()

	if err != nil {
		panic("Failed to create connection object")
	}

	go func() {
			WaitUntilConnected(connection)
			err := SendBlobData(connection, []byte("Hello Client"))

			if err != nil {
				panic(err)
			}

			for {
				select {
				case val, ok := <-connection.incoming:
					// channel has been closed
					if !ok {
						return
					}
		
					if bytes.Equal(val, []byte("Hello Server")) {
						InitiateClose(connection)
					}
				}
			}
		}
	}()

	err = UpgradeConnection(w, r, connection)

	if err != nil {
		fmt.Println("Failed to upgrade connection", err)
	}
}

```
