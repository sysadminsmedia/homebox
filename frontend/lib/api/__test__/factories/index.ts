import { faker } from "@faker-js/faker";
import { expect } from "vitest";
import { overrideParts } from "../../base/urls";
import { PublicApi } from "../../public";
import type { ItemField, LabelCreate, LocationCreate, UserRegistration } from "../../types/data-contracts";
import * as config from "../../../../test/config";
import { UserClient } from "../../user";
import { Requests } from "../../../requests";

function itemField(id = null): ItemField {
  return {
    // @ts-expect-error - not actually an issue
    id,
    name: faker.lorem.word(),
    type: "text",
    textValue: faker.lorem.sentence(),
    booleanValue: false,
    numberValue: faker.number.int(),
    timeValue: "",
  };
}

/**
 * Returns a random user registration object that can be
 * used to signup a new user.
 */
function user(): UserRegistration {
  return {
    email: faker.internet.email(),
    password: faker.internet.password(),
    name: faker.person.firstName(),
    token: "",
  };
}

function location(parentId: string | null = null): LocationCreate {
  return {
    parentId,
    name: faker.location.city(),
    description: faker.lorem.sentence(),
  };
}

function label(): LabelCreate {
  return {
    name: faker.lorem.word(),
    description: faker.lorem.sentence(),
    color: faker.color.rgb(),
  };
}

function publicClient(): PublicApi {
  overrideParts(config.BASE_URL, "/api/v1");
  const requests = new Requests("");
  return new PublicApi(requests);
}

function userClient(token: string): UserClient {
  overrideParts(config.BASE_URL, "/api/v1");
  const requests = new Requests("", token);
  return new UserClient(requests, "");
}

type TestUser = {
  client: UserClient;
  user: UserRegistration;
};

async function userSingleUse(): Promise<TestUser> {
  const usr = user();

  const pub = publicClient();
  await pub.register(usr);
  const result = await pub.login(usr.email, usr.password);

  expect(result.error).toBeFalsy();
  expect(result.status).toBe(200);

  return {
    client: new UserClient(new Requests("", result.data.token), result.data.attachmentToken),
    user: usr,
  };
}

export const factories = {
  user,
  location,
  label,
  itemField,
  client: {
    public: publicClient,
    user: userClient,
    singleUse: userSingleUse,
  },
};
