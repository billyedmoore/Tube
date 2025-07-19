package server

import (
	"encoding/binary"
	"fmt"
)

func commonEncoding(op opcode, version uint8) []byte {
	return []byte{uint8(op), version}
}

func encodeError(errorReason string) ([]byte, error) {
	blob := commonEncoding(ERROR, 0)

	maxValue := 65536 // 2^16

	length := len(errorReason)

	if length > maxValue {
		return nil, fmt.Errorf("Argument `errorReason` must be less than 2^16 characters in length.")
	}

	lengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lengthBytes, uint16(length))

	blob = append(blob, lengthBytes...)
	blob = append(blob, []byte(errorReason)...)

	return blob, nil
}

func encodeSenderAcceptance(shareCode []byte) ([]byte, error) {
	expectedShareCodeLength := 5
	if len(shareCode) != expectedShareCodeLength {
		return nil, fmt.Errorf("Argument `share_code` should be of length %d is actually of length %d.",
			expectedShareCodeLength, len(shareCode))
	}

	blob := commonEncoding(SENDER_ACCEPTED, 0)

	blob = append(blob, shareCode...)

	return blob, nil
}

func encodeRecieverAcceptance() []byte {
	return commonEncoding(RECEIVER_ACCEPTED, 0)
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
