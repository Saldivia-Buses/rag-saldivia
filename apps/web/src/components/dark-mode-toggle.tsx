"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { MoonIcon, SunIcon } from "lucide-react";

export function DarkModeToggle() {
  const [isDark, setIsDark] = useState(false);

  useEffect(() => {
    const saved = localStorage.getItem("sda-dark-mode");
    if (saved === "true") {
      document.documentElement.classList.add("dark");
      setIsDark(true);
    }
  }, []);

  const toggle = useCallback(() => {
    const next = !isDark;
    setIsDark(next);
    document.documentElement.classList.toggle("dark", next);
    localStorage.setItem("sda-dark-mode", String(next));
  }, [isDark]);

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggle}
      className="size-8"
    >
      {isDark ? (
        <SunIcon className="size-4" />
      ) : (
        <MoonIcon className="size-4" />
      )}
    </Button>
  );
}
