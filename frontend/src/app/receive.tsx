import { useState } from "react";
import { Section } from "./section";
import { Share } from "./send";

export type ReceivingShare = Share & {
  fileName?: string;
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
  onInputEntered: () => void;
}

const ReceiveInput: React.FC<ReceiveInputProps> = ({ setShare, onInputEntered }) => {
  const [inputtedShareCode, setInputtedShareCode] = useState<string>("")
  return (
    <>
      <input
        type="text"
        id="receive_share_code_input"
        className="bg-slate-500 py-2 px-4 rounded focus:bg-slate-400"
        placeholder="share_code"
        onChange={(change) => setInputtedShareCode(change.target.value)}
      />
      <button className="bg-logopink hover:bg-logopinkdark text-white font-bold py-2 px-4 rounded"
        onClick={() => { setShare({ shareCode: inputtedShareCode }); onInputEntered(); }}>
        Fetch
      </button >
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
      <p>
        {(share?.error) ? share.error : "Oooops, somthing went wrong."}
      </p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setRecieveState(ReceiveState.INPUT)}
      >
        Try again
      </button>
    </>
  );
};

const ReceiveActive: React.FC<ReceiveSubComponentProps> = ({ setRecieveState }) => {
  return (
    <>
      <p>Active</p>
      <button
        className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded"
        onClick={() => setRecieveState(ReceiveState.INPUT)}
      > Back
      </button>
    </>
  );
};


export const Recieve = () => {
  const [share, setShare] = useState<ReceivingShare | undefined>(
    undefined,
  );
  const [sendState, setReceivingState] = useState<ReceiveState>(ReceiveState.INPUT);

  let component: React.ReactNode;

  switch (sendState) {
    case ReceiveState.INPUT:
      component = (
        <ReceiveInput
          share={share}
          setShare={setShare}
          setRecieveState={setReceivingState}
          onInputEntered={() => setReceivingState(ReceiveState.ACTIVE)}
        />
      );
      break;
    case ReceiveState.ACTIVE:
      component = (
        <ReceiveActive
          share={share}
          setShare={setShare}
          setRecieveState={setReceivingState}
        />
      );
      break;
    case ReceiveState.COMPLETE:
      component = (
        <ReceiveComplete
          share={share}
          setShare={setShare}
          setRecieveState={setReceivingState}
        />
      );
      break;
    case ReceiveState.ERROR:
      component = (
        <ReceiveError
          share={share}
          setShare={setShare}
          setRecieveState={setReceivingState}
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
