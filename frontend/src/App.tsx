import { Link, Switch, Route, useLocation } from "wouter";
import {
  BrainCircuit,
  ClipboardList,
  Hammer,
  House,
  MessageSquareText,
  Settings,
  UserRound,
} from "lucide-react";
import { FC } from "react";
import { cn } from "@/lib/utils";
import Home from "./Home";
import Clipboard from "./Clipboard";
import Prompt from "./Prompt";
import LLM from "./LLM";
import User from "./User";
import Setting from "./Setting";
import Tool from "./Tool";

type IConComponent = typeof Settings;
type FeatureProps = {
  name: string;
  Icon: IConComponent;
  href: string;
};
const Feature: FC<FeatureProps> = ({ name, Icon, href }) => {
  const [location] = useLocation();
  const isActive = href === location;
  return (
    <Link
      key={name}
      href={href}
      className={cn(
        "w-8 h-8 flex items-center justify-center rounded-sm ",
        isActive
          ? "text-zinc-900 bg-white border-solid border border-zinc-200"
          : "text-zinc-600  "
      )}
    >
      <Icon size="20" />
    </Link>
  );
};

const features = [
  {
    name: "Home",
    Icon: House,
    href: "/",
  },
  {
    name: "Tools",
    Icon: Hammer,
    href: "/tools",
  },
  {
    name: "Clipboard History",
    Icon: ClipboardList,
    href: "/clipboard",
  },
  {
    name: "Prompt",
    Icon: MessageSquareText,
    href: "/prompt",
  },
  {
    name: "LLM",
    Icon: BrainCircuit,
    href: "/llm",
  },
].map((feature) => <Feature key={feature.name} {...feature} />);

const settings = [
  {
    name: "user",
    Icon: UserRound,
    href: "/user",
  },
  { name: "setting", Icon: Settings, href: "/setting" },
].map((setting) => <Feature key={setting.name} {...setting} />);

function App() {
  // const [location] = useLocation();
  return (
    <div className="w-full h-screen flex flex-col">
      {/* <div className="header h-6 flex pl-20">location: {location}</div> */}
      <div className="flex bg-zinc-200 flex-grow">
        <div className="w-10 flex flex-col justify-between  border-r border-solid border-gray-400 z-20 bg-zinc-200">
          <div className=" text-center flex flex-col items-center pt-2 gap-2">
            {features}
          </div>
          <div className="text-center flex flex-col items-center gap-2">
            {settings}
          </div>
        </div>
        <div className="flex-grow h-full ">
          <Switch>
            <Route path="/" component={Home} />
            <Route path="/tools" component={Tool} />
            <Route path="/clipboard" component={Clipboard} />
            <Route path="/prompt" component={Prompt} />
            <Route path="/llm" component={LLM} />
            <Route path="/user" component={User} />
            <Route path="/setting" component={Setting} />
          </Switch>
        </div>
      </div>
    </div>
  );
}

export default App;
