import { useDeferredValue, useState } from "react";

/**
 * Search hook using React 19 useDeferredValue for automatic debouncing.
 * The deferred value updates at lower priority, preventing API calls on every keystroke.
 */
export function useERPSearch(minLength = 2) {
  const [search, setSearch] = useState("");
  const deferredSearch = useDeferredValue(search);
  const enabled = deferredSearch.length >= minLength;
  return { search, setSearch, deferredSearch, enabled };
}
