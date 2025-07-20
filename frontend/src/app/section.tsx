import Image from "next/image";

type arrowType = "up" | "down";

interface SectionProps {
  children: React.ReactNode;
  title: string;
  colour: "pink" | "purple";
  arrow_type: arrowType;
}

interface ArrowProps {
  arrowVariant: arrowType;
}

const Arrow: React.FunctionComponent<ArrowProps> = ({ arrowVariant }) => {
  switch (arrowVariant) {
    case "up":
      return <Image src="/up.svg" alt="up arrow" width={20} height={5} />;
    case "down":
      return <Image src="/down.svg" alt="up arrow" width={20} height={10} />;
    default:
      console.warn(`Cannot draw arrow of unknown type ${arrowVariant}`);
      return <></>;
  }
};

const Section: React.FunctionComponent<SectionProps> = ({
  children,
  title,
  colour,
  arrow_type,
}) => {
  let className = `flex w-full h-full items-center justify-center`;

  if (colour == "pink") {
    className += " bg-logopinkdark";
  } else if (colour == "purple") {
    className += " bg-logopurpledark";
  } else {
    console.warn(
      'Invalid colour passed to Section should be "pink" or "purple".',
    );
  }

  return (
    <div className={className}>
      <div className="flex flex-col items-center justify-center">
        <div className="flex flex-row gap-3">
          <SectionTitle title={title} />
          <Arrow arrowVariant={arrow_type} />
        </div>
        <div className="flex flex-row gap-3 py-10 ">{children}</div>
      </div>
    </div>
  );
};

interface SectionTitleProps {
  title: string;
}

const SectionTitle: React.FunctionComponent<SectionTitleProps> = ({
  title,
}) => {
  return <p className="text-5xl">{title}</p>;
};

export const SendSection = () => {
  return (
    <Section title="SEND" colour="pink" arrow_type="up">
      <input
        type="file"
        className="bg-slate-500 file:bg-logopurple file:hover:bg-logopurpledark file:py-2 file:px-4 file:font-bold rounded "
      />
      <button className="bg-logopurple hover:bg-logopurpledark text-white font-bold py-2 px-4 rounded">
        Send
      </button>
    </Section>
  );
};

export const RecieveSection = () => {
  return (
    <Section title="RECIEVE" colour="purple" arrow_type="down">
      <input
        type="text"
        className="bg-slate-500 py-2 px-4 rounded focus:bg-slate-400"
        placeholder="share_code"
      />
      <button className="bg-logopink hover:bg-logopinkdark text-white font-bold py-2 px-4 rounded">
        Fetch
      </button>
    </Section>
  );
};
