package server

import (
	"encoding/binary"
	"fmt"
)

func assertVersion(version uint8, expectedVersion uint8) error {
	if version != expectedVersion {
		return fmt.Errorf("Protocol version {%v} is not supported", version)
	}
	return nil
}

// Takes a recieved blob and returns (opcode, protocol version, operation specific blob, error)
func commonDecoding(blob []byte) (opcode, uint8, []byte, error) {
	if len(blob) == 0 {
		return 0, 0, nil, fmt.Errorf("No data to decode")
	} else if len(blob) == 1 {
		return 0, 0, nil, fmt.Errorf("No version byte provided")
	}

	if uint8(blob[0]) > highestOpCode {
		return 0, 0, nil, fmt.Errorf("Invalid opcode provided")
	}
	return opcode(blob[0]), blob[1], blob[2:], nil
}

func decodeSenderInitiation(blob []byte) error {
	op, ver, _, err := commonDecoding(blob)

	if ver != 0 {
	}

	if err != nil {
		return err
	}

	if op != SENDER_INITIATION {
		return fmt.Errorf("Message is not a SENDER_INITIATION is a {%v}", opcode(op))
	}

	return nil
}

// Takes a recieved blob and returns (client_public_key, error)
func decodeReceiverInitiation(blob []byte) ([]byte, error) {
	op, ver, remainingBlob, err := commonDecoding(blob)

	if err != nil {
		return nil, err
	}

	err = assertVersion(ver, 0)

	if err != nil {
		return nil, err
	}

	if op != RECEIVER_INITIATION {
		return nil, fmt.Errorf("Message is not a RECEIVER_INITIATION is a {%v}", opcode(op))
	}

	publicKeyLength := 512
	if len(remainingBlob) < publicKeyLength {
		return nil, fmt.Errorf("Too few bytes (expected %v got %v)).", publicKeyLength, len(remainingBlob))
	}

	var public_key []byte = remainingBlob[:512]

	return public_key, nil
}

// Takes a recieved blob and returns (number_of_chunks, error)
func decodeMetadata(blob []byte) (uint16, error) {
	op, ver, remainingBlob, err := commonDecoding(blob)

	if err != nil {
		return 0, err
	}

	err = assertVersion(ver, 0)

	if err != nil {
		return 0, err
	}

	if op != METADATA {
		return 0, fmt.Errorf("Message is not a METADATA is a %v", opcode(op))
	}

	if len(remainingBlob) == 0 {
		return 0, fmt.Errorf("Incomplete message.")
	}

	filenameLength := remainingBlob[0]

	if len(remainingBlob) < 1 || filenameLength == 0 {
		return 0, fmt.Errorf("file_name length must be at least 1 byte.")
	}

	minNumberOfBytes := 1 + int(filenameLength) + 2

	if len(remainingBlob) < (minNumberOfBytes) {
		return 0, fmt.Errorf("Too few bytes (expected %v got %v)).", len(remainingBlob), minNumberOfBytes)
	}

	number_of_chunks_i := remainingBlob[1+int(filenameLength):]

	return binary.LittleEndian.Uint16(number_of_chunks_i), nil
}

// Takes a recieved blob and returns (chunk_number, error)
func decodeDataChunk(blob []byte) (uint16, error) {
	op, ver, remainingBlob, err := commonDecoding(blob)

	if err != nil {
		return 0, err
	}

	err = assertVersion(ver, 0)

	if err != nil {
		return 0, err
	}

	if op != DATA_CHUNK {
		return 0, fmt.Errorf("Message is not a DATA_CHUNK is a %v.", opcode(op))
	}

	if len(remainingBlob) < 4 {
		return 0, fmt.Errorf("Incomplete message.")
	}

	chunkNumberBytes := remainingBlob[:2]
	payloadLengthBytes := remainingBlob[2:4]

	if len(remainingBlob[4:]) < int(binary.LittleEndian.Uint16(payloadLengthBytes)) {
		return 0, fmt.Errorf("Incomplete message.")
	}

	chunkNumber := binary.LittleEndian.Uint16(chunkNumberBytes)

	return chunkNumber, nil
}

// Takes a recieved blob and returns (chunk_number, error)
// Chunk number of 0xFF means metadata
func decodeAcknowledge(blob []byte) (uint16, error) {
	op, ver, remainingBlob, err := commonDecoding(blob)

	if err != nil {
		return 0, err
	}

	err = assertVersion(ver, 0)

	if err != nil {
		return 0, err
	}

	if op != ACKNOWLEDGE {
		return 0, fmt.Errorf("Message is not a ACKNOWLEDGE is a %v.", opcode(op))
	}

	if len(remainingBlob) < 2 {
		return 0, fmt.Errorf("Incomplete message.")
	}

	chunkNumberBytes := remainingBlob[:2]

	chunkNumber := binary.LittleEndian.Uint16(chunkNumberBytes)

	return chunkNumber, nil
}
