import { SendSection, RecieveSection } from "./section";

export default function Home() {
  return (
    <>
      <main className="font-mono flex h-screen flex-col md:flex-row">
        <SendSection />
        <RecieveSection />
      </main>
      <footer></footer>
    </>
  );
}
