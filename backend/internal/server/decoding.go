package server

import (
	"encoding/binary"
	"fmt"
)

func assertVersion(version uint8, expected_version uint8) error {
	if version != expected_version {
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

	highest_opcode := uint8(8)
	if uint8(blob[0]) > highest_opcode {
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
	op, ver, remaining_blob, err := commonDecoding(blob)

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

	public_key_length := 512
	if len(remaining_blob) < public_key_length {
		return nil, fmt.Errorf("Too few bytes (expected %v got %v)).", public_key_length, len(remaining_blob))
	}

	var public_key []byte = remaining_blob[:512]

	return public_key, nil
}

// Takes a recieved blob and returns (number_of_chunks, error)
func decodeMetadata(blob []byte) (uint16, error) {
	op, ver, remaining_blob, err := commonDecoding(blob)

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

	if len(remaining_blob) == 0 {
		return 0, fmt.Errorf("Incomplete message.")
	}

	filename_length := remaining_blob[0]

	if len(remaining_blob) < 1 || filename_length == 0 {
		return 0, fmt.Errorf("file_name length must be at least 1 byte.")
	}

	min_numb_bytes := 1 + int(filename_length) + 2

	if len(remaining_blob) < (min_numb_bytes) {
		return 0, fmt.Errorf("Too few bytes (expected %v got %v)).", len(remaining_blob), min_numb_bytes)
	}

	number_of_chunks_i := remaining_blob[1+int(filename_length):]

	return binary.LittleEndian.Uint16(number_of_chunks_i), nil
}

// Takes a recieved blob and returns (chunk_number, error)
func decodeDataChunk(blob []byte) (uint16, error) {
	op, ver, remaining_blob, err := commonDecoding(blob)

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

	if len(remaining_blob) < 4 {
		return 0, fmt.Errorf("Incomplete message.")
	}

	chunk_number_bytes := remaining_blob[:2]
	payload_length_bytes := remaining_blob[2:4]

	if len(remaining_blob[4:]) < int(binary.LittleEndian.Uint16(payload_length_bytes)) {
		return 0, fmt.Errorf("Incomplete message.")
	}

	chunk_number := binary.LittleEndian.Uint16(chunk_number_bytes)

	return chunk_number, nil
}

// Takes a recieved blob and returns (chunk_number, error)
// Chunk number of 0xFF means metadata
func decodeAcknowledge(blob []byte) (uint16, error) {
	op, ver, remaining_blob, err := commonDecoding(blob)

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

	if len(remaining_blob) < 2 {
		return 0, fmt.Errorf("Incomplete message.")
	}

	chunk_number_bytes := remaining_blob[:2]

	chunk_number := binary.LittleEndian.Uint16(chunk_number_bytes)

	return chunk_number, nil
}
