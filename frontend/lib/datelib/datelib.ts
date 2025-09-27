import { addDays } from "date-fns";

// Always work with UTC dates internally
// Convert to local time only for display

export function format(date: Date | string): string {
  if (typeof date === "string") {
    return date;
  }
  return date.toISOString();
}

export function factorRange(offset: number = 7): [Date, Date] {
  const now = new Date();
  const start = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()));
  const end = addDays(start, offset);
  return [start, end];
}

export function factory(offset = 0): Date {
  const now = new Date();
  const date = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()));
  if (offset) {
    return addDays(date, offset);
  }
  return date;
}

export function parse(isoString: string): Date {
  return new Date(isoString);
}

// zeroTime refactored as utility for creating UTC dates without time components
export function zeroTime(date: Date): Date {
  return new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()));
}
