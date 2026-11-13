import { describe, expect, it } from "vitest";
import { filterLocations, withAssetIds } from "./use-location-helpers";

const locations = withAssetIds(
  [
    { id: "winter", name: "Winter clothes", treeString: "Bedroom > Winter clothes" },
    { id: "summer", name: "Summer clothes", treeString: "Bedroom > Summer clothes" },
    { id: "sid-in-name", name: "Shelf 000-078", treeString: "Garage > Shelf 000-078" },
  ],
  [
    { id: "winter", assetId: "000-078" },
    { id: "summer", assetId: "000-079" },
    { id: "sid-in-name", assetId: "000-080" },
  ]
);

describe("filterLocations", () => {
  it("filters locations by asset ID when the query starts with a hash", () => {
    expect(filterLocations("#000-078", locations).map(location => location.id)).toEqual(["winter"]);
  });

  it("keeps name and path search when the query has no hash", () => {
    expect(filterLocations("summer", locations).map(location => location.id)).toEqual(["summer"]);
  });

  it("keeps tree path search when the query has no hash", () => {
    expect(filterLocations("Garage", locations).map(location => location.id)).toEqual(["sid-in-name"]);
  });
});
