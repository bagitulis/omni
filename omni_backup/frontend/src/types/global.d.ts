// Global type declarations for browser APIs

// requestIdleCallback is not in all TypeScript lib definitions
interface IdleRequestOptions {
  timeout?: number;
}

interface IdleDeadline {
  didTimeout: boolean;
  timeRemaining: () => number;
}

type IdleRequestCallback = (deadline: IdleDeadline) => void;

declare function requestIdleCallback(
  callback: IdleRequestCallback,
  options?: IdleRequestOptions
): number;

declare function cancelIdleCallback(handle: number): void;
