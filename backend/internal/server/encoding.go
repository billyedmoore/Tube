package server

import "fmt"

func commonEncoding(op opcode, version uint8) []byte {
	return []byte{uint8(op), version}
}

func encodeSenderAcceptance(share_code []byte) ([]byte, error) {
	share_code_length := 5
	if len(share_code) != share_code_length {
		return nil, fmt.Errorf("Argument `share_code` should be of length %d is actually of length %d.",
			share_code_length, len(share_code))
	}

	blob := commonEncoding(SENDER_ACCEPTED, 0)

	blob = append(blob, share_code...)

	return blob, nil
}

