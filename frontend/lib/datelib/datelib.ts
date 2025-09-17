import { addDays } from "date-fns";

export function zeroTime(date: Date): Date {
  const result = new Date(date.getTime());
  result.setHours(0, 0, 0, 0);
  return result;
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
  const parts = yyyyMMdd.split("-") as [string, string, string];
  return new Date(parseInt(parts[0]), parseInt(parts[1]) - 1, parseInt(parts[2]));
}
