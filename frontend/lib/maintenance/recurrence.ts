import { MaintenancePlanIntervalUnit } from "~~/lib/api/types/data-contracts";

function addInterval(base: Date, intervalValue: number, intervalUnit: MaintenancePlanIntervalUnit): Date {
  const next = new Date(base);
  switch (intervalUnit) {
    case MaintenancePlanIntervalUnit.Hour:
      next.setHours(next.getHours() + intervalValue);
      return next;
    case MaintenancePlanIntervalUnit.Day:
      next.setDate(next.getDate() + intervalValue);
      return next;
    case MaintenancePlanIntervalUnit.Week:
      next.setDate(next.getDate() + intervalValue * 7);
      return next;
    case MaintenancePlanIntervalUnit.Month: {
      const y = next.getUTCFullYear();
      const m = next.getUTCMonth();
      const day = next.getUTCDate();
      const totalMonths = y * 12 + m + intervalValue;
      const ty = Math.floor(totalMonths / 12);
      const tm = totalMonths - ty * 12;
      const lastDay = new Date(Date.UTC(ty, tm + 1, 0)).getUTCDate();
      next.setUTCFullYear(ty, tm, Math.min(day, lastDay));
      return next;
    }
    case MaintenancePlanIntervalUnit.Year: {
      const y = next.getUTCFullYear() + intervalValue;
      const m = next.getUTCMonth();
      const day = next.getUTCDate();
      const lastDay = new Date(Date.UTC(y, m + 1, 0)).getUTCDate();
      next.setUTCFullYear(y, m, Math.min(day, lastDay));
      return next;
    }
    default:
      return next;
  }
}

export function getNextNDueDates(
  firstDueDate: Date,
  intervalValue: number,
  intervalUnit: MaintenancePlanIntervalUnit,
  count: number
): Date[] {
  const result: Date[] = [];
  let current = new Date(firstDueDate);
  for (let i = 0; i < count; i++) {
    result.push(new Date(current));
    current = addInterval(current, intervalValue, intervalUnit);
  }
  return result;
}
