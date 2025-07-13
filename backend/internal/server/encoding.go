package server

import "fmt"

func commonEncoding(op opcode, version uint8) []byte {
	return []byte{uint8(op), version}
}

func encodeSenderAcceptance(shareCode []byte) ([]byte, error) {
	shareCodeLength := 5
	if len(shareCode) != shareCodeLength {
		return nil, fmt.Errorf("Argument `share_code` should be of length %d is actually of length %d.",
			shareCodeLength, len(shareCode))
	}

	blob := commonEncoding(SENDER_ACCEPTED, 0)

	blob = append(blob, shareCode...)

	return blob, nil
}

func encodeRecieverAcceptance() []byte {
	return commonEncoding(RECIEVER_ACCEPTED, 0)
}

func encodeReady(publicKey []byte) ([]byte, error) {
	publicKeyLength := 512
	actualPublicKeyLength := len(publicKey)
	if actualPublicKeyLength != publicKeyLength {
		err := fmt.Errorf("Public key should be %d bytes is actually %d.",
			publicKeyLength, actualPublicKeyLength)
		return nil, err
	}

	blob := commonEncoding(READY, 0)

	blob = append(blob, publicKey...)

	return blob, nil
}
