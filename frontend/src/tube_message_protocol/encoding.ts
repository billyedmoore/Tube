import { encodeKey } from "../crypto";
import { CHUNK_SIZE, OpCode } from "./types";

export const encodeReceiverInitiation = async (
  publicKey: CryptoKey,
): Promise<ArrayBuffer> => {
  const opcode = OpCode.RECEIVER_INITIATION;
  const ver = 0;
  const key = await encodeKey(publicKey);
  console.log(key);
  const keyBuff = new Uint8Array(key);
  const messageBuffer = new Uint8Array(keyBuff.byteLength + 2);

  messageBuffer[0] = opcode;
  messageBuffer[1] = ver;
  messageBuffer.set(keyBuff, 2);

  return messageBuffer.buffer;
};

export const encodeMetadata = (file: File): ArrayBuffer => {
  const opcode = OpCode.METADATA;
  const ver = 0;
  const encoder = new TextEncoder();
  const file_name = encoder.encode(file.name);
  const file_length = file.size;
  const number_chunks = Math.ceil(file_length / CHUNK_SIZE);

  const messageBuffer = new ArrayBuffer(file_name.length + 5);
  const view = new DataView(messageBuffer);

  view.setUint8(0, opcode);
  view.setUint8(1, ver);
  view.setUint8(2, file_name.byteLength);
  const fileNameView = new Uint8Array(messageBuffer, 3, file_name.byteLength);
  fileNameView.set(file_name);
  view.setUint16(file_length + 3, number_chunks);

  return messageBuffer;
};
