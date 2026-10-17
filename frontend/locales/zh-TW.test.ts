import { IntlMessageFormat } from "intl-messageformat";
import { expect, test } from "vitest";

import messages from "./zh-TW.json";

test("label generator tips are valid ICU messages", () => {
  for (const message of [
    messages.reports.label_generator.tip_1,
    messages.reports.label_generator.tip_2,
    messages.reports.label_generator.tip_3,
  ]) {
    expect(() =>
      new IntlMessageFormat(message, "zh-TW").format(),
    ).not.toThrow();
  }
});
