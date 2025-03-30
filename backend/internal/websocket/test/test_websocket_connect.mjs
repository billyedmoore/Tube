async function connectAndServe(serverUrl) {
	const ws = new WebSocket(serverUrl);

	ws.onopen = () => {
		// 15 second timout
		setTimeout(() => { ws.close(); Deno.exit() }, 200)
	};

	ws.onmessage = (event) => {
		const decoder = new TextDecoder();
		event.data.text().then((text) => console.log("Message from server:", text))
	};

	ws.onerror = (error) => {
		throw error
	};

	ws.onclose = () => {
		console.log("Closed")
	}

}

const serverUrl = 'ws://localhost:8080/ws';
connectAndServe(serverUrl);
