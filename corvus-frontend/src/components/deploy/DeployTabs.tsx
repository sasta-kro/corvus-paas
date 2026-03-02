import type { ReactNode } from "react";
import * as Tabs from "@radix-ui/react-tabs";

interface DeployTabsProps {
  activeTab: string;
  onTabChange: (tab: string) => void;
  children: ReactNode;
}

export default function DeployTabs({ activeTab, onTabChange, children }: DeployTabsProps) {
  const tabStyle = (value: string): React.CSSProperties => ({
    color: activeTab === value ? "var(--sumi)" : "var(--sumi-wash)",
    fontFamily: '"EB Garamond", serif',
    fontWeight: 700,
    fontSize: "0.95rem",
    letterSpacing: "0.02em",
    background: activeTab === value ? "var(--paper-warm)" : "transparent",
    borderRadius: "3px 3px 0 0",
    borderBottom: "none",
    boxShadow: activeTab === value
      ? "inset 0 -3px 0 0 var(--sumi), 0 2px 0 0 var(--paper-warm)"
      : "none",
  });

  return (
    <Tabs.Root value={activeTab} onValueChange={onTabChange}>
      <Tabs.List className="flex gap-1 mb-6">
        {[
          { value: "quick", label: "Quick Deploy" },
          { value: "zip", label: "Zip Upload" },
          { value: "github", label: "GitHub Repo" },
        ].map((tab) => (
          <Tabs.Trigger
            key={tab.value}
            value={tab.value}
            className="ink-tab px-4 py-2.5 cursor-pointer"
            style={tabStyle(tab.value)}
          >
            {tab.label}
          </Tabs.Trigger>
        ))}
      </Tabs.List>
      <div className="brush-divider-thin mb-6" />
      {children}
    </Tabs.Root>
  );
}

export { Tabs };
