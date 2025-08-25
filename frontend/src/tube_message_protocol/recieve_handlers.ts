import { ReceiveState, ReceivingShare } from "../app/receive";
import { decodeReceiverAccepted, decodeMetadata } from "./decoding";
import { assertBlobMessageType } from "./send_handlers";

export interface ReceiveMessageHandler {
  (
    ws: WebSocket,
    incoming: MessageEvent,
    share: ReceivingShare | undefined,
    setSendState: React.Dispatch<React.SetStateAction<ReceiveState>>,
  ): Promise<[ReceivingShare | undefined, ReceiveMessageHandler]>;
}

const handleNoFurtherMessages: ReceiveMessageHandler = async (
  _ws: WebSocket,
  _event: MessageEvent,
  _share: ReceivingShare | undefined,
  _setState: React.Dispatch<React.SetStateAction<ReceiveState>>,
) => {
  throw new Error("Unexpected message.");
};

export const handleReceiverAccepted: ReceiveMessageHandler = async (
  _ws: WebSocket,
  incoming: MessageEvent,
  share: ReceivingShare | undefined,
  _setSendState: React.Dispatch<React.SetStateAction<ReceiveState>>,
) => {
  console.log("HandlingRecieverAccepted");

  assertBlobMessageType(incoming);
  const payload = await incoming.data.arrayBuffer();
  decodeReceiverAccepted(payload);

  return [share, handleMetadata];
};

export const handleMetadata: ReceiveMessageHandler = async (
  _: WebSocket,
  incoming: MessageEvent,
  share: ReceivingShare | undefined,
  _setSendState: React.Dispatch<React.SetStateAction<ReceiveState>>,
) => {
  console.log("HandlingMetaData");
  assertBlobMessageType(incoming);
  const payload = await incoming.data.arrayBuffer();
  const key = share?.keys?.privateKey;
  if (!key) {
    throw new Error("Key must be set to decode MetaData");
  }
  !decodeMetadata(payload, key);

  // TODO: HandleDataChunk
  return [share, handleNoFurtherMessages];
};
