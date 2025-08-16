"use client";

import { Recieve } from "./receive";
import { Send } from "./send";

export default function Home() {
  return (
    <>
      <main className="font-mono flex h-screen flex-col md:flex-row">
        <Send />
        <Recieve />
      </main>
      <footer></footer>
    </>
  );
}
