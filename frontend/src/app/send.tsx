import { useState, useRef, useEffect } from "react";
import { Section } from "./section";
import { Mutex } from "../tube_message_protocol/mutex";
import {
  handleSenderAccepted,
  SendMessageHandler,
} from "../tube_message_protocol/send_handlers";

// Shared between the send and receive sides of the share
export type Share = {
  shareCode?: string;
  error?: string;
};

export type SendingShare = Share & {
  recieverPublicKey?: CryptoKey;
  file?: File;
  ws?: WebSocket;
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

interface SendInputProps extends SendSubComponentProps {
  onInputEntered: (file: File) => void;
}

const SendInput: React.FC<SendInputProps> = ({ share, onInputEntered }) => {
  const [file, setFile] = useState<File | undefined>(undefined);
  return (
    <>
      <input
        type="file"
        id="send_file_picker"
        className="bg-slate-500 file:bg-logopurple file:hover:bg-logopurpledark file:py-2 file:px-4 file:font-bold rounded"
        onChange={(event) => setFile(event.target.files?.[0] || undefined)}
      />
      <button
        className={
          "bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        }
        onClick={() => {
          if (share != undefined) {
            console.error("Share should not yet exist, but is not undefined.");
          }

          if (file) {
            console.log(file);
            onInputEntered(file);
          } else {
            window.alert("Please select a file.");
          }
        }}
      >
        Send
      </button>
    </>
  );
};

const SendWaiting: React.FC<SendSubComponentProps> = ({ share }) => {
  return (
    <>
      <p>
        ShareCode: <b className="text-xl font-mono">{share?.shareCode}</b>
      </p>
    </>
  );
};

const SendActive: React.FC<SendSubComponentProps> = ({ setSendState }) => {
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

const SendComplete: React.FC<SendSubComponentProps> = ({ setSendState }) => {
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
  setSendState,
}) => {
  return (
    <>
      <p>{share?.error ? share.error : "Oooops, somthing went wrong."}</p>
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
  const [share, setShare] = useState<SendingShare | undefined>(undefined);
  const [sendState, setSendingState] = useState<SendState>(SendState.INPUT);

  // This means the race condition is way less likely to occur
  // since on every re-render this will be updated sychronusly
  // Really I think share should be a useReducer value maybe
  // with the sendState also encapsulated.
  const shareRef = useRef(share);

  useEffect(() => {
    shareRef.current = share;
  }, [share]);

  const startShareConnection = (file: File) => {
    let ws: WebSocket;
    ws = new WebSocket("ws://localhost:8080/send");

    const incomingMessageMutex = new Mutex();
    ws.onopen = () => {
      setShare({ file });
      // TODO: change this to an encodingFunction
      ws.send(Uint8Array.from([1, 0]).buffer);
    };

    ws.onerror = (error) => {
      console.error("SEND WebSocket connection failed - ", error);
      setShare({ ...share, error: "Backend Connection Failed" });
      setSendingState(SendState.ERROR);
    };

    let handler: SendMessageHandler = handleSenderAccepted;
    ws.onmessage = async (message) => {
      try {
        await incomingMessageMutex.lock();
        let newShare: SendingShare | undefined;
        [newShare, handler] = await handler(
          ws,
          message,
          shareRef.current,
          setSendingState,
        );
        setShare(newShare);
      } catch (e) {
        console.error("Error encountered by handler - ", e);
        const err = e instanceof Error ? e.message : "";
        setShare({ ...share, error: err });
        setSendingState(SendState.ERROR);
      } finally {
        incomingMessageMutex.unlock();
      }
    };
  };

  let component: React.ReactNode;

  switch (sendState) {
    case SendState.INPUT:
      component = (
        <SendInput
          share={share}
          setShare={setShare}
          setSendState={setSendingState}
          onInputEntered={startShareConnection}
        />
      );
      break;
    case SendState.WAITING:
      component = (
        <SendWaiting
          share={share}
          setShare={setShare}
          setSendState={setSendingState}
        />
      );
      break;
    case SendState.ACTIVE:
      component = (
        <SendActive
          share={share}
          setShare={setShare}
          setSendState={setSendingState}
        />
      );
      break;
    case SendState.COMPLETE:
      component = (
        <SendComplete
          share={share}
          setShare={setShare}
          setSendState={setSendingState}
        />
      );
    case SendState.ERROR:
      component = (
        <SendError
          share={share}
          setShare={setShare}
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
