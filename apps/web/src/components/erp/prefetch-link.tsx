"use client";

import Link, { type LinkProps } from "next/link";
import { useRef } from "react";
import {
  type FetchQueryOptions,
  useQueryClient,
} from "@tanstack/react-query";

interface PrefetchLinkProps<TData>
  extends Omit<LinkProps, "onMouseEnter" | "onFocus" | "onTouchStart"> {
  prefetch: () => FetchQueryOptions<TData>;
  children: React.ReactNode;
  className?: string;
}

export function PrefetchLink<TData>({
  prefetch,
  children,
  className,
  ...linkProps
}: PrefetchLinkProps<TData>) {
  const queryClient = useQueryClient();
  const prefetched = useRef(false);

  const trigger = () => {
    if (prefetched.current) return;
    prefetched.current = true;
    queryClient.prefetchQuery(prefetch());
  };

  return (
    <Link
      {...linkProps}
      className={className}
      onMouseEnter={trigger}
      onFocus={trigger}
      onTouchStart={trigger}
    >
      {children}
    </Link>
  );
}
