async function connectAndServe(serverUrl) {
	const ws = new WebSocket(serverUrl);

	ws.onmessage = (event) => {
		// This is used by the test to check the client recieved the correct data
		event.data.text().then((text) => console.log("Message from server:", text))
		const encoder = new TextEncoder()
		// This tests the server can recieve data
		// Strings are unsupported in this library so we use bytes
		ws.send(encoder.encode("Recieved"))
	};

	ws.onerror = (error) => {
		throw error
	};
}

const serverUrl = 'ws://localhost:8080/ws';
connectAndServe(serverUrl);
