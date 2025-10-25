import { FC, HTMLAttributes, useState, useEffect } from "react";
import { cn } from "./lib/utils";
import { GetDirs } from "../wailsjs/go/hub/Model";
import { hub } from "../wailsjs/go/models";

const toolTemplate = (dirs: hub.Dirs) => ({
  id: 123,
  name: "find_file_by_name",
  description: "find file by name",
  parameters: {
    type: "object",
    properties: {
      dir: {
        type: "string",
      },
      name: {
        type: "string",
      },
      required: ["dir", "name"],
    },
  },
  type: "cmd_line_tool",
  wd: dirs.home,
  cmd: ["find", "$dir", "-type", "f", "-iname", "'$name'"],
  env: {},
  timeout: "30s",
  isStream: false,
  error: "",
  dependencies: [],
  testcases: [],
  concurrencyGroupID: 1,
  concurrencyGroup: {
    id: 1,
    name: "default",
    description: "default concurrency group",
    maxConcurrency: 5,
  },
});

export const Editor: FC<HTMLAttributes<HTMLDivElement>> = ({ className }) => {
  const [code, setCode] = useState<string>("");
  useEffect(() => {
    GetDirs().then((dirs) => {
      setCode(encodeURIComponent(JSON.stringify(toolTemplate(dirs), null, 2)));
    });
  }, []);
  return (
    <div className={cn("w-full h-[calc(100vh-3rem)] flex flex-col", className)}>
      <iframe
        src={`/monaco/index.html?code=${code}&language=json&theme=vs-light`}
        title="Monaco Editor"
        className="flex-1 border-none h-full"
        allow="clipboard-read; clipboard-write"
      />
    </div>
  );
};
