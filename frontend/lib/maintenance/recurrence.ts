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
    case MaintenancePlanIntervalUnit.Month:
      next.setMonth(next.getMonth() + intervalValue);
      return next;
    case MaintenancePlanIntervalUnit.Year:
      next.setFullYear(next.getFullYear() + intervalValue);
      return next;
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
