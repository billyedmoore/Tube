# Backend

## Websockets

> [!WARNING]
> This is in early development. Information on this page may be out of date or factually incorrect.

Tube uses a custom websocket library for websocket communication.

Tube's websocket library is based on go's "net" and "net/http" internally.

### Usage Basics

To import:

```go
import ("github.com/billyedmoore/tube/websocket")
```

Create a connection with:
```go
connection,err := websocket.CreateConnection()

if err != nil {
  // There was an error and the connection wasn't created
}
```

The connection isn't active until it is upgraded. 
Example usage can be seen below:
```go
func handler(w http.ResponseWriter, r *http.Request){
  connection,err := websocket.CreateConnection()

  if err != nil {
   // There was an error and the connection wasn't created
   // You can use w to send an error to the client
  }
  
  err = UpgradeConnection(w,r,connection)

  if err != nil {
    // There was an error the connection may or may not have been
    // successfully Hijacked so you should still try to send an 
    // error over http. An effort may be made in future to ensure
    // client can still be notified if failure occurs later.
  }
  
  for {
    select {
    case data, ok := <- connection.incoming:
      if !ok {
        // Channel closed, signals the connection has been closed.
        return
        }
      fmt.Println("Recieved :" data)
    }
  }
}

```
