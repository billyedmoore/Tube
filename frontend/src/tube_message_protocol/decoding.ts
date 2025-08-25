import { decodeKey, decrypt } from "../crypto";
import {
  Ready,
  SenderAccepted,
  CommonParsedTubeMessage,
  OpCode,
  ShareCodeLength as SHARE_CODE_LENGTH,
} from "./types";

const decodeCommon = (messageBlob: ArrayBuffer): CommonParsedTubeMessage => {
  const dataView = new DataView(messageBlob);

  if (messageBlob.byteLength < 2) {
    throw new Error("The message is too short to be valid.");
  }
  const opcode: number = dataView.getInt8(0);
  const ver: number = dataView.getInt8(1);

  const remainingBlob = messageBlob.slice(2);

  return { opcode, ver, remainingBlob };
};

export const decodeSenderAccepted = (
  messageBlob: ArrayBuffer,
): SenderAccepted => {
  let { opcode, ver, remainingBlob } = decodeCommon(messageBlob);

  if (ver != 0) {
    throw new Error("Unsupported version.");
  }

  if (opcode != OpCode.SENDER_ACCEPTED) {
    throw new Error(
      `Message of unexpected type - ${opcode} (expected OpCode.SENDER_ACCEPTED).`,
    );
  }

  if (remainingBlob.byteLength < SHARE_CODE_LENGTH) {
    throw new Error("Share code too short.");
  }

  return { shareCode: remainingBlob.slice(0, SHARE_CODE_LENGTH) };
};

// Reciever accepted contains no information, just may be used to ensure receiver
// initiation is recieved in future.
export const decodeReceiverAccepted = (messageBlob: ArrayBuffer): void => {
  let { opcode, ver } = decodeCommon(messageBlob);

  if (ver != 0) {
    throw new Error("Unsupported version.");
  }

  if (opcode != OpCode.RECEIVER_ACCEPTED) {
    throw new Error(
      `Message of unexpected type - ${opcode} (expected OpCode.RECEIVER_ACCEPTED).`,
    );
  }
};

export const decodeReady = async (messageBlob: ArrayBuffer): Promise<Ready> => {
  let { opcode, ver, remainingBlob } = decodeCommon(messageBlob);

  if (ver != 0) {
    throw new Error("Unsupported version.");
  }

  if (opcode != OpCode.READY) {
    throw new Error(
      `Message of unexpected type - ${opcode} (expected OpCode.READY).`,
    );
  }

  /*
  if (remainingBlob.byteLength < 550) {
    throw new Error("Client public key should be 512 bytes.");
  }*/

  // THIS IS THE PROBLEM
  console.log(remainingBlob.slice(0, 550));
  const receiverKey = await decodeKey(remainingBlob.slice(0, 550));

  return { receiverPublicKey: receiverKey };
};

export const decodeMetadata = async (
  messageBlob: ArrayBuffer,
  privateKey: CryptoKey,
): Promise<Metadata> => {
  let { opcode, ver, remainingBlob } = decodeCommon(messageBlob);

  if (ver != 0) {
    throw new Error("Unsupported version.");
  }

  if (opcode != OpCode.METADATA) {
    throw new Error(
      `Message of unexpected type - ${opcode} (expected OpCode.METADATA).`,
    );
  }
  const view = new Uint8Array(remainingBlob);

  const filenameLength = view[0];

  if (remainingBlob.byteLength < filenameLength + 3) {
    throw new Error(
      `MetaData message not long enough to contain an ${filenameLength} filename as claimed.`,
    );
  }

  const fileNameBuff = await decrypt(
    privateKey,
    remainingBlob.slice(1, filenameLength + 1),
  );

  const nChunksLowByte = view[filenameLength + 1];
  const nChunksHighByte = view[filenameLength + 2];

  const nChunks = (nChunksHighByte << 8) | nChunksLowByte;

  remainingBlob.slice(filenameLength + 1, filenameLength + 3);

  const decoder = new TextDecoder("utf-8");

  return { fileName: decoder.decode(fileNameBuff), numberOfChunks: nChunks };
};
