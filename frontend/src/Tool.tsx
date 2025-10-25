import {
  SidebarProvider,
  SidebarTrigger,
  useSidebar,
} from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { SquareTerminal } from "lucide-react";
import { cn } from "./lib/utils";
import { Editor } from "./Editor";

export default function Tool() {
  // Menu items.
  const items = [
    {
      title: "Dashboard",
      url: "#",
      icon: SquareTerminal,
    },
    {
      title: "Inbox",
      url: "#",
      icon: SquareTerminal,
    },
    {
      title: "Calendar",
      url: "#",
      icon: SquareTerminal,
    },
    {
      title: "Search",
      url: "#",
      icon: SquareTerminal,
    },
    {
      title: "Settings",
      url: "#",
      icon: SquareTerminal,
    },
  ];

  return (
    <SidebarProvider className="h-[calc(100vh-1.5rem)] !min-h-0 flex">
      <AppSidebar items={items} title="tool list" className="" />
      <Main className="flex-grow" />
    </SidebarProvider>
  );
}

type MainProps = React.HTMLAttributes<HTMLDivElement>;
function Main({ className }: MainProps) {
  const { open } = useSidebar();
  return (
    <main
      className={cn(
        className,
        "h-svh border-solid border-zinc-400 flex flex-col",
        open && "border-l"
      )}
    >
      <div className="h-8 bg-zinc-200 border-b border-solid border-zinc-400">
        {!open ? (
          <SidebarTrigger className="" />
        ) : (
          <div className="h-[28px] " />
        )}
      </div>
      <div className="flex-grow bg-white p-2">
        <Editor className="w-1/2 min-w-[400px]" />
      </div>
    </main>
  );
}
