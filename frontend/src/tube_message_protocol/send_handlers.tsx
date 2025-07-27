import { SendingShare, SendState } from "../app/send";

interface MessageHandler {
  (
    ws: WebSocket,
    incoming: MessageEvent,
    share: SendingShare | undefined,
    setShare: React.Dispatch<React.SetStateAction<SendingShare | undefined>>,
    setSendState: React.Dispatch<React.SetStateAction<SendState>>,
  ): Promise<MessageHandler>;
}

const base64Encode = (blob: ArrayBuffer): string => {
  const blobAsUint8 = new Uint8Array(blob);
  let binaryString = "";
  for (let i = 0; i < blobAsUint8.byteLength; i++) {
    binaryString += String.fromCharCode(blobAsUint8[i]);
  }

  return btoa(binaryString);
};

const handleNoFurtherMessages: MessageHandler = (
  _: WebSocket,
  __: MessageEvent,
  ___: SendingShare | undefined,
  ____: React.Dispatch<React.SetStateAction<SendingShare | undefined>>,
  _____: React.Dispatch<React.SetStateAction<SendState>>,
) => {
  throw new Error("Unexpected message.");
};

const handleSenderAccepted: MessageHandler = async (
  _: WebSocket,
  incoming: MessageEvent,
  share: SendingShare | undefined,
  setShare: React.Dispatch<React.SetStateAction<SendingShare | undefined>>,
  setSendState: React.Dispatch<React.SetStateAction<SendState>>,
) => {
  if (!(incoming.data instanceof ArrayBuffer)) {
    const newShare = {
      ...share,
      error: "Recevied non ArrayBuffer websocket message",
    };
    setShare(newShare);
  }
  try {
    const { shareCode } = decodeSenderAccepted(incoming.data);

    const newShare = { ...share, shareCode: base64Encode(shareCode) };
    setShare(newShare);

    setSendState(SendState.WAITING);
  } catch (e) {
    let err_msg = "Something went wrong.";
    if (e instanceof Error) {
      err_msg = e.message;
    }
    const newShare = { ...share, error: err_msg };
    setShare(newShare);
  }

  return handleNoFurtherMessages;
};
