"use client";

import Link, { type LinkProps } from "next/link";
import { useRef } from "react";
import {
  type FetchQueryOptions,
  useQueryClient,
} from "@tanstack/react-query";

interface PrefetchLinkProps<TData>
  extends Omit<LinkProps, "onMouseEnter" | "onFocus" | "onTouchStart" | "prefetch"> {
  prefetchQuery: () => FetchQueryOptions<TData>;
  children: React.ReactNode;
  className?: string;
}

export function PrefetchLink<TData>({
  prefetchQuery,
  children,
  className,
  ...linkProps
}: PrefetchLinkProps<TData>) {
  const queryClient = useQueryClient();
  const prefetched = useRef(false);

  const trigger = () => {
    if (prefetched.current) return;
    prefetched.current = true;
    queryClient.prefetchQuery(prefetchQuery());
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
