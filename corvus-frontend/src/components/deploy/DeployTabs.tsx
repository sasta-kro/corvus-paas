import type { ReactNode } from "react";
import * as Tabs from "@radix-ui/react-tabs";

interface DeployTabsProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  children: ReactNode;
}

/** Tab navigation wrapper using Radix UI Tabs */
export default function DeployTabs({
  activeTab,
  onTabChange,
  children,
}: DeployTabsProps) {
  return (
    <Tabs.Root value={activeTab} onValueChange={onTabChange}>
      <Tabs.List className="flex border-b border-gray-200 mb-6">
        <Tabs.Trigger
          value="quick"
          className="px-4 py-2.5 text-sm font-medium text-gray-500 hover:text-black border-b-2 border-transparent data-[state=active]:border-black data-[state=active]:text-black transition-colors cursor-pointer"
        >
          Quick Deploy
        </Tabs.Trigger>
        <Tabs.Trigger
          value="zip"
          className="px-4 py-2.5 text-sm font-medium text-gray-500 hover:text-black border-b-2 border-transparent data-[state=active]:border-black data-[state=active]:text-black transition-colors cursor-pointer"
        >
          Zip Upload
        </Tabs.Trigger>
        <Tabs.Trigger
          value="github"
          className="px-4 py-2.5 text-sm font-medium text-gray-500 hover:text-black border-b-2 border-transparent data-[state=active]:border-black data-[state=active]:text-black transition-colors cursor-pointer"
        >
          GitHub Repo
        </Tabs.Trigger>
      </Tabs.List>
      {children}
    </Tabs.Root>
  );
}

export { Tabs };

