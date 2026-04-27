import { describe, expect, it } from "vitest";
import { MaintenancePlanIntervalUnit } from "~~/lib/api/types/data-contracts";
import { getNextNDueDates } from "./recurrence";

describe("getNextNDueDates", () => {
  it("includes scheduled date as the first due date", () => {
    const scheduledDate = new Date("2026-03-10T09:30:00Z");
    const dueDates = getNextNDueDates(scheduledDate, 1, MaintenancePlanIntervalUnit.Week, 3);

    expect(dueDates).toHaveLength(3);
    const [firstDueDate, secondDueDate, thirdDueDate] = dueDates;
    expect(firstDueDate?.toISOString()).toBe("2026-03-10T09:30:00.000Z");
    expect(secondDueDate?.toISOString()).toBe("2026-03-17T09:30:00.000Z");
    expect(thirdDueDate?.toISOString()).toBe("2026-03-24T09:30:00.000Z");
  });

  it("applies month interval consistently", () => {
    const scheduledDate = new Date("2026-01-15T00:00:00Z");
    const dueDates = getNextNDueDates(scheduledDate, 1, MaintenancePlanIntervalUnit.Month, 3);

    const [firstDueDate, secondDueDate, thirdDueDate] = dueDates;
    expect(firstDueDate?.toISOString()).toBe("2026-01-15T00:00:00.000Z");
    expect(secondDueDate?.toISOString()).toBe("2026-02-15T00:00:00.000Z");
    expect(thirdDueDate?.toISOString()).toBe("2026-03-15T00:00:00.000Z");
  });
});
