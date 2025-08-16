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
      return <Image src="/up.svg" alt="up arrow" width={20} height={10} />;
    case "down":
      return <Image src="/down.svg" alt="up arrow" width={20} height={10} />;
    default:
      console.warn(`Cannot draw arrow of unknown type ${arrowVariant}`);
      return <></>;
  }
};

export const Section: React.FunctionComponent<SectionProps> = ({
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
      <div className="flex flex-col items-center gap-10 justify-center">
        <div className="flex flex-row gap-3">
          <SectionTitle title={title} />
          <Arrow arrowVariant={arrow_type} />
        </div>
        <div className="flex flex-row gap-3 md:shrink w-full items-center">
          {children}
        </div>
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
