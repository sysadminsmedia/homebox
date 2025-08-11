import { addDays } from "date-fns";

/*
 * Formats a date as a string
 * */
export function format(date: Date | string): string {
  if (typeof date === "string") {
    return date;
  }
  return date.toISOString().split("T")[0];
}

export function zeroTime(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

export function factorRange(offset: number = 7): [Date, Date] {
  const date = zeroTime(new Date());

  return [date, addDays(date, offset)];
}

export function factory(offset = 0): Date {
  if (offset) {
    return addDays(zeroTime(new Date()), offset);
  }

  return zeroTime(new Date());
}

export function parse(yyyyMMdd: string): Date {
  const parts = yyyyMMdd.split("-");
  return new Date(parseInt(parts[0]), parseInt(parts[1]) - 1, parseInt(parts[2]));
}
