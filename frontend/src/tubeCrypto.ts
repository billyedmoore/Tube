namespace tubeCrypto {

	const encryptionAlgoName: string = "RSA-OAEP"
	const encryptionKeyLength: number = 4096 // 512 bytes
	const encryptionHash: string = "SHA-512"

	export async function generateKeyPair(): Promise<CryptoKeyPair> {
		return (
			crypto.subtle.generateKey({
				name: encryptionAlgoName,
				modulusLength: encryptionKeyLength,
				publicExponent: new Uint8Array([1, 0, 1]), // default value
				hash: encryptionHash
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
			{ name: encryptionAlgoName, hash: encryptionHash },
			false, // Shouldn't need to be exported again
			["encrypt"])
	}

	export async function encrypt(publicKey: CryptoKey, data: ArrayBuffer): Promise<ArrayBuffer> {
		return crypto.subtle.encrypt({ name: encryptionAlgoName }, publicKey, data)
	}

	export async function decrypt(privateKey: CryptoKey, data: ArrayBuffer): Promise<ArrayBuffer> {
		return crypto.subtle.decrypt({ name: encryptionAlgoName }, privateKey, data)
	}
}

export default tubeCrypto
