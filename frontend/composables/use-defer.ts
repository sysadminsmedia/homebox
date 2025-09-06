/* eslint-disable @typescript-eslint/no-explicit-any */
type DeferFunction<TArgs extends any[], TReturn> = (...args: TArgs) => TReturn;

// useDefer is a function that takes a function and returns a function that
// calls the original function and then calls the onComplete function.
export function useDefer<TArgs extends any[], TReturn>(
  onComplete: (...args: TArgs) => void,
  func: DeferFunction<TArgs, TReturn>
): DeferFunction<TArgs, TReturn> {
  return (...args: TArgs) => {
    let result: TReturn;
    try {
      result = func(...args);
    } finally {
      onComplete(...args);
    }

    return result;
  };
}
