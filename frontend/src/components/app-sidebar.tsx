import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { FC, HTMLAttributes } from "react";

type AppSidebarProps = HTMLAttributes<HTMLDivElement> & {
  title: string;
  items: {
    title: string;
    url: string;
    icon: FC<any>;
  }[];
};
export function AppSidebar({ items, title }: AppSidebarProps) {
  return (
    <Sidebar variant="inset" className="left-10  p-0 bg-zinc-400">
      <SidebarContent className="bg-zinc-100">
        <SidebarGroup className="p-0">
          <SidebarGroupLabel className="flex justify-between bg-zinc-200 border-b border-solid border-zinc-400 rounded-none pl-4">
            <div>{title}</div>
            <SidebarTrigger />
          </SidebarGroupLabel>
          <SidebarGroupContent className="px-2">
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <a href={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </a>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  );
}
