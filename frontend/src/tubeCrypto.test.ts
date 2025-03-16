import tubeCrypto from "./tubeCrypto"

async function testEncryptDecrypt(data: ArrayBuffer): Promise<ArrayBuffer> {
	const keys = await tubeCrypto.generateKeyPair();
	const blob = await tubeCrypto.encrypt(keys.publicKey, data)
	return tubeCrypto.decrypt(keys.privateKey, blob)
}

async function testEncodeDecodeKey(): Promise<string> {
	const keys = await tubeCrypto.generateKeyPair();
	const pubKeyBlob = await tubeCrypto.encodeKey(keys.publicKey)
	const pubKey = await tubeCrypto.decodeKey(pubKeyBlob)
	const blob = await tubeCrypto.encrypt(pubKey, new TextEncoder().encode("Hello World!").buffer)
	const decryptedBlob = await tubeCrypto.decrypt(keys.privateKey, blob)
	return new TextDecoder().decode(new Uint8Array(decryptedBlob))
}
test("encrypt-decrypt test", () => {
	const data: ArrayBuffer = new TextEncoder().encode("Hello World!").buffer;
	testEncryptDecrypt(data).then((decryptedData) => {
		expect(new Uint8Array(data)).toEqual(new Uint8Array(decryptedData))
	})
}
)

test("encode-decode test", () => {
	testEncodeDecodeKey().then(txt => expect(txt).toEqual("Hello World!"))
})


