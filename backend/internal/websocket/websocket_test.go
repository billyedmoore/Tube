package websocket

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
	"testing"
	"time"
)

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server failed to stop\n")
	}
	fmt.Printf("Server stopped\n")
}

func testWebsocketConnectHandler(w http.ResponseWriter, r *http.Request) {
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
			case val, ok := <-connection.Incoming:
				// channel has been closed
				if !ok {
					return
				}

				if bytes.Equal(val, []byte("Recieved")) {
					//Close out the connection
					fmt.Println("All good closing")
					InitiateClose(connection)
				}
			}
		}

	}()

	err = UpgradeConnection(w, r, connection)

	if err != nil {
		fmt.Println("Failed to upgrade connection", err)
	}

}

func TestWebsocketConnect(t *testing.T) {

	var wg sync.WaitGroup
	errors := make(chan error, 5)
	defer close(errors)

	http.HandleFunc("/ws", testWebsocketConnectHandler)
	server := &http.Server{
		Addr: ":8080",
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.ListenAndServe()

		if err != http.ErrServerClosed {
			errors <- err
		}
	}()

	wg.Add(1)
	go func(timeout int) {
		defer shutdownServer(server)
		defer wg.Done()

		// backup for the timeout in the js
		ctx, cancel := context.WithTimeout(context.Background(), (time.Duration(timeout) * time.Second))
		defer cancel()

		println("deno run test/test_websocket_connect.mjs")

		cmd := exec.CommandContext(ctx, "deno", "run", "--allow-net=localhost", "test/test_websocket_connect.mjs")

		output, err := cmd.CombinedOutput()
		fmt.Printf("\nSTDOUT: %s \nSTDERR: %e\n", strconv.Quote(string(output)), err)

		if ctx.Err() == context.DeadlineExceeded {
			errors <- fmt.Errorf("Client server run timeout after %d seconds.", timeout)
			return
		}

		if err != nil {
			errors <- fmt.Errorf("Client server run failed : %s - %s ", output, err.Error())
			return
		}

		if string(output) != "Message from server: Hello Client\n" {
			errors <- fmt.Errorf("Didn't get the expected output.")
			return
		}

	}(7)

	wg.Wait()

	select {
	case err, ok := <-errors:
		if !ok {
			fmt.Println("Errors channel is closed")
		} else {
			t.Error(err)
		}
	default:
		fmt.Println("Ran with no errors")
	}
}
