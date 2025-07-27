import { useEffect, useState } from "react";
import { Section } from "./section";
import { Mutex } from "../tube_message_protocol/mutex";

// Shared between the send and receive sides of the share
export type Share = {
  shareCode?: string;
  error?: string;
};

export type SendingShare = Share & {
  // Stuff special to the Share side of the Share
};

export enum SendState {
  INPUT = "input",
  WAITING = "waiting",
  ACTIVE = "active",
  COMPLETE = "complete",
  ERROR = "error",
}

interface SendSubComponentProps {
  share: SendingShare | undefined;
  setShare: React.Dispatch<React.SetStateAction<SendingShare | undefined>>;
  setSendState: React.Dispatch<React.SetStateAction<SendState>>;
}

const SendInput: React.FC<SendSubComponentProps> = ({
  share,
  setShare,
  setSendState,
}) => {
  return (
    <>
      <input
        type="file"
        id="send_file_picker"
        className="bg-slate-500 file:bg-logopurple file:hover:bg-logopurpledark file:py-2 file:px-4 file:font-bold rounded"
      />
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setSendState(SendState.WAITING)}
      >
        Send
      </button>
    </>
  );
};

const SendWaiting: React.FC<SendSubComponentProps> = ({
  share,
  setShare,
  setSendState,
}) => {
  return (
    <>
      <p>Waiting.</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setSendState(SendState.ACTIVE)}
      >
        Send
      </button>
    </>
  );
};

const SendActive: React.FC<SendSubComponentProps> = ({
  share,
  setShare,
  setSendState,
}) => {
  return (
    <>
      <p>Active.</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setSendState(SendState.COMPLETE)}
      >
        Send
      </button>
    </>
  );
};

const SendComplete: React.FC<SendSubComponentProps> = ({
  share,
  setShare,
  setSendState,
}) => {
  return (
    <>
      <p>Complete.</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setSendState(SendState.INPUT)}
      >
        Send
      </button>
    </>
  );
};

const SendError: React.FC<SendSubComponentProps> = ({
  share,
  setShare,
  setSendState,
}) => {
  return (
    <>
      <p>Oooops, somthing went wrong.</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setSendState(SendState.INPUT)}
      >
        Start new share
      </button>
    </>
  );
};

export const Send = () => {
  const [openShare, setOpenShare] = useState<SendingShare | undefined>(
    undefined,
  );
  const [sendState, setSendingState] = useState<SendState>(SendState.INPUT);

  const incomingMessageMutex = Mutex();
  useEffect(() => {
    // TODO: move this to an env variable
    const ws = new WebSocket("ws://localhost:8080/send");
    let handler = (ws.onopen = (_) => {
      // Send Initation
    });

    ws.onmessage = async (message) => {
      // I think we can skip the mutex by making the handler functions non-asyc
      // Can check back and see if the async is needed
      await incomingMessageMutex.lock();
      handler = await handler(
        ws,
        message,
        openShare,
        setOpenShare,
        sendState,
        setSendingState,
      );
      incomingMessageMutex.unlock();
    };
  });

  let component: React.ReactNode;

  switch (sendState) {
    case SendState.INPUT:
      component = (
        <SendInput
          share={openShare}
          setShare={setOpenShare}
          setSendState={setSendingState}
        />
      );
      break;
    case SendState.WAITING:
      component = (
        <SendWaiting
          share={openShare}
          setShare={setOpenShare}
          setSendState={setSendingState}
        />
      );
      break;
    case SendState.ACTIVE:
      component = (
        <SendActive
          share={openShare}
          setShare={setOpenShare}
          setSendState={setSendingState}
        />
      );
      break;
    case SendState.COMPLETE:
      component = (
        <SendComplete
          share={openShare}
          setShare={setOpenShare}
          setSendState={setSendingState}
        />
      );
    case SendState.ERROR:
      component = (
        <SendError
          share={openShare}
          setShare={setOpenShare}
          setSendState={setSendingState}
        />
      );
      break;
  }
  return (
    <Section title="SEND" colour="pink" arrow_type="up">
      {component}
    </Section>
  );
};
