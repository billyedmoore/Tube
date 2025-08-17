const ENCRYPTION_ALGORITHM: string = "RSA-OAEP"
const KEY_LENGTH: number = 4096 // 512 bytes
const ENCRYPTION_HASH: string = "SHA-512"

export async function generateKeyPair(): Promise<CryptoKeyPair> {
	return (
		crypto.subtle.generateKey({
			name: ENCRYPTION_ALGORITHM,
			modulusLength: KEY_LENGTH,
			publicExponent: new Uint8Array([1, 0, 1]), // default value
			hash: ENCRYPTION_HASH
		},
			true,
			["encrypt", "decrypt"])
	) as Promise<CryptoKeyPair>
}

export async function encodeKey(key: CryptoKey): Promise<ArrayBuffer> {
	return crypto.subtle.exportKey("spki", key)
}

export async function decodeKey(encodedKey: ArrayBuffer): Promise<CryptoKey> {
	// Only to be used on Public Keys from the other client
	return crypto.subtle.importKey("spki",
		encodedKey,
		{ name: ENCRYPTION_ALGORITHM, hash: ENCRYPTION_HASH },
		false, // Shouldn't need to be exported again
		["encrypt"])
}

export async function encrypt(publicKey: CryptoKey, data: ArrayBuffer): Promise<ArrayBuffer> {
	return crypto.subtle.encrypt({ name: ENCRYPTION_ALGORITHM }, publicKey, data)
}

export async function decrypt(privateKey: CryptoKey, data: ArrayBuffer): Promise<ArrayBuffer> {
	return crypto.subtle.decrypt({ name: ENCRYPTION_ALGORITHM }, privateKey, data)
}
