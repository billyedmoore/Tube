import { useState } from "react";
import { Section } from "./section";

// Shared between the send and receive sides of the share
type Share = {
  share_code: string;
};

type SendingShare = Share & {
  // Stuff special to the Share side of the Share
};

enum SendState {
  INPUT = "input",
  WAITING = "waiting",
  ACTIVE = "active",
  COMPLETE = "complete",
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

export const Send = () => {
  const [openShare, setOpenShare] = useState<SendingShare | undefined>(
    undefined,
  );
  const [sendState, setSendingState] = useState<SendState>(SendState.INPUT);

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
      break;
  }
  return (
    <Section title="SEND" colour="pink" arrow_type="up">
      {component}
    </Section>
  );
};
