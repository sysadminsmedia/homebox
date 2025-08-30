import { faker } from "@faker-js/faker";
import { beforeAll, describe, expect, test } from "vitest";
import type { UserClient } from "../../user";
import { factories } from "../factories";

type ImportObj = {
  [`HB.import_ref`]: string;
  [`HB.location`]: string;
  [`HB.labels`]: string;
  [`HB.quantity`]: number;
  [`HB.name`]: string;
  [`HB.description`]: string;
  [`HB.insured`]: boolean;
  [`HB.serial_number`]: string;
  [`HB.model_number`]: string;
  [`HB.manufacturer`]: string;
  [`HB.notes`]: string;
  [`HB.purchase_price`]: number;
  [`HB.purchase_from`]: string;
  [`HB.purchase_time`]: string;
  [`HB.lifetime_warranty`]: boolean;
  [`HB.warranty_expires`]: string;
  [`HB.warranty_details`]: string;
  [`HB.sold_to`]: string;
  [`HB.sold_price`]: number;
  [`HB.sold_time`]: string;
  [`HB.sold_notes`]: string;
};

function toCsv(data: ImportObj[]): string {
  const headers = Object.keys(data[0]!).join("\t");
  const rows = data.map(row => {
    return Object.values(row).join("\t");
  });
  return [headers, ...rows].join("\n");
}

function importFileGenerator(entries: number): ImportObj[] {
  const imports: Partial<ImportObj>[] = [];

  const pick = (arr: string[]) => arr[Math.floor(Math.random() * arr.length)];

  const labels = faker.word.words(5).split(" ").join(";");
  const locations = faker.word.words(3).split(" ");

  const half = Math.floor(entries / 2);

  // YYYY-MM-DD
  const formatDate = (date: Date) => date.toISOString().split("T")[0];

  for (let i = 0; i < entries; i++) {
    imports.push({
      [`HB.import_ref`]: faker.database.mongodbObjectId(),
      [`HB.location`]: pick(locations),
      [`HB.labels`]: labels,
      [`HB.quantity`]: Number(faker.number.int(2)),
      [`HB.name`]: faker.word.words(3),
      [`HB.description`]: "",
      [`HB.insured`]: faker.datatype.boolean(),
      [`HB.serial_number`]: faker.string.alphanumeric(5),
      [`HB.model_number`]: faker.string.alphanumeric(5),
      [`HB.manufacturer`]: faker.string.alphanumeric(5),
      [`HB.notes`]: "",
      [`HB.purchase_from`]: faker.person.fullName(),
      [`HB.purchase_price`]: faker.number.int(100),
      [`HB.purchase_time`]: faker.date.past().toDateString(),
      [`HB.lifetime_warranty`]: half > i,
      [`HB.warranty_details`]: "",
      [`HB.sold_to`]: faker.person.fullName(),
      [`HB.sold_price`]: faker.number.int(100),
      [`HB.sold_time`]: formatDate(faker.date.past()),
      [`HB.sold_notes`]: "",
    });
  }

  return imports as ImportObj[];
}

describe("group related statistics tests", () => {
  const TOTAL_ITEMS = 30;
  const labelData: Record<string, number> = {};
  const locationData: Record<string, number> = {};

  let tAPI: UserClient | undefined;
  const imports = importFileGenerator(TOTAL_ITEMS);

  const api = (): UserClient => {
    if (!tAPI) {
      throw new Error("API not initialized");
    }
    return tAPI;
  };

  beforeAll(async () => {
    // -- Setup --
    const { client } = await factories.client.singleUse();
    tAPI = client;

    const csv = toCsv(imports);

    const setupResp = await client.items.import(new Blob([csv], { type: "text/csv" }));

    expect(setupResp.status).toBe(204);

    for (const item of imports) {
      const labels = item[`HB.labels`].split(";");
      for (const label of labels) {
        if (labelData[label]) {
          labelData[label] += item[`HB.purchase_price`];
        } else {
          labelData[label] = item[`HB.purchase_price`];
        }
      }

      const location = item[`HB.location`];
      if (locationData[location]) {
        locationData[location] += item[`HB.purchase_price`];
      } else {
        locationData[location] = item[`HB.purchase_price`];
      }
    }
  });

  test("Validate Group Statistics", async () => {
    const { status, data } = await api().stats.group();
    expect(status).toBe(200);

    expect(data.totalItems).toEqual(TOTAL_ITEMS);
    expect(data.totalLabels).toEqual(11); // default + new
    expect(data.totalLocations).toEqual(11); // default + new
    expect(data.totalUsers).toEqual(1);
    expect(data.totalWithWarranty).toEqual(Math.floor(TOTAL_ITEMS / 2));
  });

  test("Validate Labels Statistics", async () => {
    const { status, data } = await api().stats.labels();
    expect(status).toBe(200);

    for (const label of data) {
      expect(label.total).toEqual(labelData[label.name]);
    }
  });

  test("Validate Locations Statistics", async () => {
    const { status, data } = await api().stats.locations();
    expect(status).toBe(200);

    for (const location of data) {
      expect(location.total).toEqual(locationData[location.name]);
    }
  });

  test("Validate Purchase Over Time", async () => {
    const { status, data } = await api().stats.totalPriceOverTime();
    expect(status).toBe(200);
    expect(data.entries.length).toEqual(TOTAL_ITEMS);
  });
});
