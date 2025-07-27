enum OpCode {
  SENDER_INITIATION = 1,
  SENDER_ACCEPTED = 2,
  RECEIVER_INITIATION = 3,
  RECEIVER_ACCEPTED = 4,
  READY = 5,
  METADATA = 6,
  DATA_CHUNK = 7,
  ACKNOWLEDGE = 8,
  ERROR = 9,
}

interface CommonPrelude {
  opcode: OpCode;
  ver: number;
  remainingBlob: ArrayBuffer;
}

interface SenderAccepted {
  shareCode: ArrayBuffer;
}

const decodeCommon = (messageBlob: ArrayBuffer): CommonPrelude => {
  const dataView = new DataView(messageBlob);

  if (messageBlob.byteLength < 2) {
    throw new Error("The message is too short to be valid.");
  }
  const opcode: number = dataView.getInt8(0);
  const ver: number = dataView.getInt8(0);

  const remainingBlob = messageBlob.slice(2);

  return { opcode, ver, remainingBlob };
};

const decodeSenderAccepted = (messageBlob: ArrayBuffer) => {
  let { opcode, ver, remainingBlob } = decodeCommon(messageBlob);

  if (ver != 0) {
    throw new Error("Unsupported version.");
  }

  if (opcode != OpCode.SENDER_ACCEPTED) {
    throw new Error(
      `Message of unexpected type - ${opcode} (expected OpCode.SENDER_ACCEPTED).`,
    );
  }

  if (remainingBlob.byteLength < 5) {
    throw new Error("Share code too short.");
  }

  return { shareCode: remainingBlob };
};
