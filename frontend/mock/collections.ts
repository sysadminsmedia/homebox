export type Collection = { id: string; name: string };

export type User = {
  id: string;
  name: string;
  email: string;
  created_at?: string;
  role: "admin" | "user" | string;
  password_set?: boolean;
  oidc_set?: boolean;
  collections: {
    id: string;
    role: "owner" | "admin" | "editor" | "viewer";
  }[];
};

export const collections: Collection[] = [
  { id: "c1", name: "Personal Inventory" },
  { id: "c2", name: "Office Equipment" },
  { id: "c3", name: "Workshop Tools" },
];

export const users: User[] = [
  {
    id: "1",
    name: "Alice Admin",
    email: "alice@example.com",
    created_at: new Date(new Date().setFullYear(new Date().getFullYear() - 2)).toISOString(),
    role: "admin",
    password_set: true,
    collections: [
      { id: collections[0]!.id, role: "owner" },
      { id: collections[1]!.id, role: "admin" },
      { id: collections[2]!.id, role: "editor" },
    ],
  },
  {
    id: "2",
    name: "Bob User",
    email: "bob@example.com",
    created_at: new Date(new Date().setFullYear(new Date().getFullYear() - 1)).toISOString(),
    role: "user",
    password_set: true,
    oidc_set: true,
    collections: [
      { id: collections[1]!.id, role: "owner" },
      { id: collections[2]!.id, role: "admin" },
    ],
  },
  {
    id: "3",
    name: "Charlie",
    email: "charlie@example.com",
    created_at: new Date().toISOString(),
    role: "user",
    password_set: false,
    // collections[3] was out of range (only 0..2 exist). Use collections[2].
    collections: [{ id: collections[2]!.id, role: "owner" }],
  },
];

export type Invite = {
  id: string;
  collectionId: string;
  role?: "owner" | "admin" | "editor" | "viewer";
  created_at?: string;
  expires_at?: string;
  max_uses?: number;
  uses?: number;
};

export const invites: Invite[] = [
  {
    id: "i1",
    collectionId: collections[0]!.id,
    role: "viewer",
    created_at: new Date().toISOString(),
    expires_at: new Date(new Date().setDate(new Date().getDate() + 7)).toISOString(),
    max_uses: 5,
    uses: 2,
  },
];

// Simple in-memory fake API operating on the above arrays.
export const api = {
  // by is the person who is requesting the collections, include the number of members and their role
  getCollections(by: string = "1") {
    const user = users.find(u => u.id === by);
    if (!user) return [];
    return user.collections
      .map(c => {
        const collection = collections.find(col => col.id === c.id);
        if (!collection) return null;
        // find number of people with access to this collection
        const count = users.reduce((acc, u) => {
          const hasAccess = u.collections.some(uc => uc.id === collection.id);
          return acc + (hasAccess ? 1 : 0);
        }, 0);
        return {
          ...collection,
          count,
          role: c.role,
        };
      })
      .filter(Boolean);
  },
  getUsers(): User[] {
    return users;
  },
  getInvites(): Invite[] {
    return invites;
  },
  getUser(id: string) {
    return users.find(u => u.id === id);
  },
  addUser(input: Partial<User>) {
    const u: User = {
      id: input.id ?? String(Date.now()),
      name: input.name ?? "",
      email: input.email ?? "",
      role: input.role ?? "user",
      password_set: input.password_set ?? false,
      oidc_set: input.oidc_set ?? false,
      collections: input.collections ?? [],
    };
    users.unshift(u);
    return u;
  },
  updateUser(updated: User) {
    const idx = users.findIndex(u => u.id === updated.id);
    if (idx >= 0) users.splice(idx, 1, { ...updated });
    return updated;
  },
  deleteUser(id: string) {
    const idx = users.findIndex(u => u.id === id);
    if (idx >= 0) {
      users.splice(idx, 1);
      return true;
    }
    return false;
  },
  addInvite(input: Partial<Invite>) {
    const inv: Invite = {
      id: input.id ?? `i${Date.now()}`,
      collectionId: input.collectionId ?? collections[0]!.id,
      role: input.role ?? "viewer",
      created_at: new Date().toISOString(),
      expires_at: input.expires_at ? input.expires_at : undefined,
      max_uses: input.max_uses ? input.max_uses : undefined,
      uses: 0,
    };
    invites.unshift(inv);
    return inv;
  },
  deleteInvite(id: string) {
    const idx = invites.findIndex(i => i.id === id);
    if (idx >= 0) invites.splice(idx, 1);
    return idx >= 0;
  },
  addCollection(input: Partial<Collection>) {
    const col: Collection = { id: input.id ?? `c${Date.now()}`, name: input.name ?? "New Collection" };
    collections.push(col);
    // add user[0] to collection
    users[0]!.collections.push({ id: col.id, role: "owner" });
    return col;
  },
  updateCollection(updated: Collection) {
    const idx = collections.findIndex(c => c.id === updated.id);
    if (idx >= 0) collections.splice(idx, 1, { ...updated });
    return updated;
  },
  addUserToCollection(userId: string, collectionId: string, role: "owner" | "admin" | "editor" | "viewer" = "viewer") {
    const u = users.find(x => x.id === userId);
    if (!u) return null;
    const exists = u.collections.find(c => c.id === collectionId);
    if (exists) {
      exists.role = role;
      return exists;
    }
    const mem = { id: collectionId, role } as { id: string; role: "owner" | "admin" | "editor" | "viewer" };
    u.collections.push(mem);
    return mem;
  },
  removeUserFromCollection(userId: string, collectionId: string) {
    const u = users.find(x => x.id === userId);
    if (!u) return false;
    const idx = u.collections.findIndex(c => c.id === collectionId);
    if (idx >= 0) {
      const wasOwner = u.collections[idx]!.role === "owner";
      u.collections.splice(idx, 1);
      // if removed owner, and no other owners exist for that collection, delete the collection
      if (wasOwner) {
        const stillOwner = users.some(other =>
          (other.collections ?? []).some(c => c.id === collectionId && c.role === "owner")
        );
        if (!stillOwner) {
          // remove collection
          const cidx = collections.findIndex(c => c.id === collectionId);
          if (cidx >= 0) collections.splice(cidx, 1);
          // remove membership from all users
          users.forEach(mu => {
            mu.collections = (mu.collections ?? []).filter(c => c.id !== collectionId);
          });
          // remove invites to that collection
          for (let i = invites.length - 1; i >= 0; i--) {
            if (invites[i]!.collectionId === collectionId) invites.splice(i, 1);
          }
        }
      }
      return true;
    }
    return false;
  },
};

export default { collections, users, invites, api };
