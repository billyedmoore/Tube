import { useState } from "react";
import { Section } from "./section";
import { Share } from "./send";
import { Mutex } from "../tube_message_protocol/mutex";
import { generateKeyPair } from "../crypto";
import { encodeReceiverInitiation } from "../tube_message_protocol/encoding";
import {
  handleReceiverAccepted,
  ReceiveMessageHandler,
} from "@/tube_message_protocol/recieve_handlers";

export type ReceivingShare = Share & {
  fileName?: string;
  keys?: CryptoKeyPair;
};

export enum ReceiveState {
  INPUT = "input",
  ACTIVE = "active",
  COMPLETE = "complete",
  ERROR = "error",
}

interface ReceiveSubComponentProps {
  share: ReceivingShare | undefined;
  setShare: React.Dispatch<React.SetStateAction<ReceivingShare | undefined>>;
  setRecieveState: React.Dispatch<React.SetStateAction<ReceiveState>>;
}

interface ReceiveInputProps extends ReceiveSubComponentProps {
  onInputEntered: (share: ReceivingShare) => void;
}

const ReceiveInput: React.FC<ReceiveInputProps> = ({ onInputEntered }) => {
  const [inputtedShareCode, setInputtedShareCode] = useState<string>("");
  return (
    <>
      <input
        type="text"
        id="receive_share_code_input"
        className="bg-slate-500 py-2 px-4 rounded focus:bg-slate-400"
        placeholder="share_code"
        onChange={(change) => setInputtedShareCode(change.target.value)}
      />
      <button
        className="bg-logopink hover:bg-logopinkdark text-white font-bold py-2 px-4 rounded"
        onClick={() => {
          onInputEntered({ shareCode: inputtedShareCode });
        }}
      >
        Fetch
      </button>
    </>
  );
};

const ReceiveComplete: React.FC<ReceiveSubComponentProps> = ({
  setRecieveState,
}) => {
  return (
    <>
      <p>Complete.</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setRecieveState(ReceiveState.INPUT)}
      >
        Recive somthing else
      </button>
    </>
  );
};

const ReceiveError: React.FC<ReceiveSubComponentProps> = ({
  share,
  setRecieveState,
}) => {
  return (
    <>
      <p>{share?.error ? share.error : "Oooops, somthing went wrong."}</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setRecieveState(ReceiveState.INPUT)}
      >
        Try again
      </button>
    </>
  );
};

const ReceiveActive: React.FC<ReceiveSubComponentProps> = ({
  setRecieveState,
}) => {
  return (
    <>
      <p>Active</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setRecieveState(ReceiveState.INPUT)}
      >
        {" "}
        Back
      </button>
    </>
  );
};

export const Recieve = () => {
  const [share, setShare] = useState<ReceivingShare | undefined>(undefined);
  const [receiveState, setReceiveState] = useState<ReceiveState>(
    ReceiveState.INPUT,
  );

  const startRecieveConnection = (share: ReceivingShare) => {
    let ws: WebSocket;
    ws = new WebSocket(
      `ws://localhost:8080/receive?share_code=${share?.shareCode}`,
    );

    ws.onopen = async () => {
      const keyPair = await generateKeyPair();
      const receiverInitiation = await encodeReceiverInitiation(
        keyPair.publicKey,
      );
      ws.send(receiverInitiation);
      setShare({ ...share, keys: keyPair });
    };

    ws.onerror = () => {
      setShare({ ...share, error: "Backend Connection Failed" });
      setReceiveState(ReceiveState.ERROR);
    };

    const incomingMessageMutex = new Mutex();
    let handler: ReceiveMessageHandler = handleReceiverAccepted;
    ws.onmessage = async (message) => {
      await incomingMessageMutex.lock();
      handler = await handler(ws, message, share, setShare, setReceiveState);
      incomingMessageMutex.unlock();
    };

    setShare({ ...share });
  };

  let component: React.ReactNode;

  switch (receiveState) {
    case ReceiveState.INPUT:
      component = (
        <ReceiveInput
          share={share}
          setShare={setShare}
          setRecieveState={setReceiveState}
          onInputEntered={startRecieveConnection}
        />
      );
      break;
    case ReceiveState.ACTIVE:
      component = (
        <ReceiveActive
          share={share}
          setShare={setShare}
          setRecieveState={setReceiveState}
        />
      );
      break;
    case ReceiveState.COMPLETE:
      component = (
        <ReceiveComplete
          share={share}
          setShare={setShare}
          setRecieveState={setReceiveState}
        />
      );
      break;
    case ReceiveState.ERROR:
      component = (
        <ReceiveError
          share={share}
          setShare={setShare}
          setRecieveState={setReceiveState}
        />
      );
      break;
  }

  return (
    <Section title="RECIEVE" colour="purple" arrow_type="down">
      {component}
    </Section>
  );
};
