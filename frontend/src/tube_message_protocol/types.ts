export enum OpCode {
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

export const ShareCodeLength = 6;

export interface CommonParsedTubeMessage {
  opcode: OpCode;
  ver: number;
  remainingBlob: ArrayBuffer;
}

export interface SenderAccepted {
  shareCode: ArrayBuffer;
}

export interface Ready {
  receiverPublicKey: CryptoKey;
}

export interface Metadata {
  fileName: string;
  fileSize: number;
  numberOfChunks: number;
}

// Put this somewhere more sensible
export const CHUNK_SIZE = 128000;
