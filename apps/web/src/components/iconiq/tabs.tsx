"use client";

import { motion, LayoutGroup } from "framer-motion";

const DEFAULT_TABS = ["Overview", "Integrations", "Activity", "Settings"];

export type TabsVariant = "spotlight" | "underline" | "chip";

const slideTransition = {
  type: "spring" as const,
  stiffness: 500,
  damping: 35,
  mass: 0.8,
};

export interface AnimatedTabsProps {
  /** Visual style: "spotlight" | "underline" | "chip". */
  variant: TabsVariant;
  /** Currently active tab value. */
  value: string;
  /** Called when the user selects a tab. */
  onValueChange: (value: string) => void;
  /** Optional list of tab labels. Defaults to Overview, Integrations, Activity, Settings. */
  tabs?: string[];
  /** Optional className for the wrapper. */
  className?: string;
}

export default function AnimatedTabs({
  variant,
  value,
  onValueChange,
  tabs = DEFAULT_TABS,
  className,
}: AnimatedTabsProps) {
  return (
    <LayoutGroup id={variant}>
      <div
        className={`relative flex items-center ${
          variant === "spotlight"
            ? "gap-1 p-1 rounded-full border border-border bg-background"
            : variant === "underline"
              ? "border-b border-border gap-6"
              : "gap-1.5"
        } ${className ?? ""}`}
      >
        {tabs.map((tab) => {
          const isActive = value === tab;
          return (
            <button
              key={tab}
              onClick={() => onValueChange(tab)}
              type="button"
              className={`relative cursor-pointer text-sm font-medium transition-colors duration-200 outline-none select-none ${
                variant === "spotlight"
                  ? "px-5 py-2"
                  : variant === "underline"
                    ? "px-4 py-2 pb-3"
                    : "px-5 py-2"
              } ${
                variant === "chip" && isActive
                  ? "text-background"
                  : isActive
                    ? "text-foreground"
                    : "text-muted-foreground hover:text-foreground/80"
              }`}
            >
              {/* Spotlight: outlined capsule that slides */}
              {isActive && variant === "spotlight" && (
                <motion.span
                  layoutId="tab-indicator"
                  className="absolute inset-0 z-0 rounded-full bg-foreground/6 ring-1 ring-foreground/20"
                  transition={slideTransition}
                  style={{ originY: "top" }}
                />
              )}

              {/* Underline */}
              {isActive && variant === "underline" && (
                <motion.span
                  layoutId="tab-indicator"
                  className="absolute bottom-0 left-0 right-0 h-[2px] bg-foreground rounded-full"
                  transition={slideTransition}
                />
              )}

              {/* Chip: solid dark inverted background */}
              {isActive && variant === "chip" && (
                <motion.span
                  layoutId="tab-indicator"
                  className="absolute inset-0 z-0 rounded-full bg-foreground shadow-[0_2px_8px_rgba(0,0,0,0.15)]"
                  transition={slideTransition}
                />
              )}

              {/* Hover state for inactive chip tabs */}
              {!isActive && variant === "chip" && (
                <span className="absolute inset-0 z-0 rounded-full bg-transparent hover:bg-muted transition-colors duration-200" />
              )}

              {/* Hover state for inactive spotlight tabs */}
              {!isActive && variant === "spotlight" && (
                <span className="absolute inset-0 z-0 rounded-full bg-transparent hover:bg-foreground/3 transition-colors duration-200" />
              )}

              <span className="relative z-10 select-none">{tab}</span>
            </button>
          );
        })}
      </div>
    </LayoutGroup>
  );
}
