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

func encodeRecieverAcceptance() []byte {
	return commonEncoding(RECIEVER_ACCEPTED, 0)
}

func encodeReady(pub_key []byte) ([]byte, error) {
	pub_key_length := 512
	actual_pub_key_length := len(pub_key)
	if actual_pub_key_length != pub_key_length {
		err := fmt.Errorf("Public key should be %d bytes is actually %d.",
			pub_key_length, actual_pub_key_length)
		return nil, err
	}

	blob := commonEncoding(READY, 0)

	blob = append(blob, pub_key...)

	return blob, nil
}
