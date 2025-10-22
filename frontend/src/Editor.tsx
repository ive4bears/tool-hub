import { FC } from "react";
// @ts-ignore
import { initVimMode } from "monaco-vim";

export const Editor: FC = () => {
  return (
    <div className="w-full h-[calc(100vh-32px)] flex flex-col">
      <iframe
        src="/monaco/index.html?code=initial_code&language=javascript&theme=vs-dark"
        title="Monaco Editor"
        className="flex-1 border-none w-[50%] h-full"
        allow="clipboard-read; clipboard-write"
      />
    </div>
  );
};
