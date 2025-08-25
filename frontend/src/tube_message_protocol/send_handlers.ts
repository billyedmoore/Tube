import { SendingShare, SendState } from "../app/send";
import { decodeSenderAccepted, decodeReady } from "./decoding";
import { encodeMetadata } from "./encoding";

export interface SendMessageHandler {
  (
    ws: WebSocket,
    incoming: MessageEvent,
    share: SendingShare | undefined,
    setSendState: React.Dispatch<React.SetStateAction<SendState>>,
  ): Promise<[SendingShare | undefined, SendMessageHandler]>;
}

export const assertBlobMessageType = (message: MessageEvent) => {
  if (!(message.data instanceof Blob)) {
    throw new Error(
      "SERVER ERROR - Recevied non ArrayBuffer websocket message",
    );
  }
};

/*
 * Handler works as a state machine, the next returned handler handling the next message.
 */

const base64Encode = (blob: ArrayBuffer): string => {
  const blobAsUint8 = new Uint8Array(blob);
  console.log(blobAsUint8);
  let binaryString = "";
  for (let i = 0; i < blobAsUint8.byteLength; i++) {
    binaryString += String.fromCharCode(blobAsUint8[i]);
  }

  return btoa(binaryString);
};

const handleNoFurtherMessages: SendMessageHandler = async (
  _ws: WebSocket,
  _event: MessageEvent,
  _share: SendingShare | undefined,
  _setState: React.Dispatch<React.SetStateAction<SendState>>,
) => {
  throw new Error("Unexpected message.");
};

export const handleSenderAccepted: SendMessageHandler = async (
  _: WebSocket,
  incoming: MessageEvent,
  share: SendingShare | undefined,
  setSendState: React.Dispatch<React.SetStateAction<SendState>>,
) => {
  console.log("HandleSenderAccepted");
  assertBlobMessageType(incoming);
  const payload = await incoming.data.arrayBuffer();
  const { shareCode } = decodeSenderAccepted(payload); // Throws if invalid

  const newShare = { ...share, shareCode: base64Encode(shareCode) };

  setSendState(SendState.WAITING);

  return [newShare, handleReady];
};

const handleReady: SendMessageHandler = async (
  ws: WebSocket,
  incoming: MessageEvent,
  share: SendingShare | undefined,
  setSendState: React.Dispatch<React.SetStateAction<SendState>>,
) => {
  console.log("HandleReady");
  assertBlobMessageType(incoming);
  const payload = await incoming.data.arrayBuffer();
  const { receiverPublicKey } = await decodeReady(payload); // Throws if invalid

  const newShare: SendingShare = {
    ...share,
    recieverPublicKey: receiverPublicKey,
  };

  if (!share?.file) {
    throw new Error("CLIENT ERROR - File not set, nothing to send.");
  }
  {
    ws.send(encodeMetadata(share?.file));
  }

  setSendState(SendState.ACTIVE);

  return [newShare, handleNoFurtherMessages];
};
