interface FileRecieveSession {
	sendID: string;
	senderPublicKey: string;
}

function handleRecieveFormComplete(event: SubmitEvent): void {
	event.preventDefault();

	const fileName = (document.getElementById('fileName') as HTMLInputElement).value;
	const sendID = (document.getElementById('sendID') as HTMLInputElement).value;

	console.log(`Saving the share ${sendID} to ${fileName}.`)
}
