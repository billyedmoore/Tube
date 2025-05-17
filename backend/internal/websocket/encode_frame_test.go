package websocket

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestEncodeBinaryFrame(t *testing.T) {
	frm, err := newBinaryFrame([]byte("Hello Client"))

	if err != nil {
		t.Fatalf("Failed to create frame")
	}

	data, err := encodeFrame(frm)

	if err != nil {
		t.Fatal("Failed to encode frame")
	}

	wanted := []byte{0x82, 0x0C, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x20, 0x43, 0x6C, 0x69, 0x65, 0x6E, 0x74}

	if !bytes.Equal(data, wanted) {
		t.Errorf("Data should be %v but is %v", wanted, data)
	}

}

func TestDecodeBinaryFrame(t *testing.T) {
	frm, err := newBinaryFrame([]byte("Hello Client"))

	if err != nil {
		t.Fatal("Failed to create frame")
	}

	data, err := encodeFrame(frm)

	fmt.Printf("frm: %v", frm)

	if err != nil {
		t.Fatal("Failed to encode frame")
	}

	newFrame, err := decodeFrame(data)

	if err != nil {
		t.Fatalf("Failed to decode frame %v", err)
	}

	if !reflect.DeepEqual(frm, newFrame) {
		t.Errorf("Decoded frame doesnt match, original=%v newFrame=%v\n", frm, newFrame)
	}

}
