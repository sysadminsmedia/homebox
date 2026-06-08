import type { TemplateField } from "./api/types/data-contracts";

export const NIL_UUID = "00000000-0000-0000-0000-000000000000";
export const DEFAULT_TEMPLATE_FIELD_TIME_VALUE = "1970-01-01T00:00:00.000Z";

export function newTemplateField(): TemplateField {
  return {
    id: NIL_UUID,
    name: "",
    type: "text",
    booleanValue: false,
    numberValue: 0,
    textValue: "",
    timeValue: DEFAULT_TEMPLATE_FIELD_TIME_VALUE,
  };
}
