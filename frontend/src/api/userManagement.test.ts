import { describe, it, expect, vi, beforeEach } from "vitest";

const { mockGet, mockPost, mockPatch, mockDelete } = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
  mockPatch: vi.fn(),
  mockDelete: vi.fn(),
}));

vi.mock("./client", () => ({
  default: {
    get: mockGet,
    post: mockPost,
    patch: mockPatch,
    delete: mockDelete,
  },
}));

import {
  listAllUsers,
  createUser,
  updateUser,
  deleteUser,
  changePassword,
} from "./userManagement";

describe("listAllUsers", () => {
  beforeEach(() => vi.clearAllMocks());

  it("returns list of users on success", async () => {
    const users = [
      {
        id: "u1",
        tenant_id: "t1",
        username: "alice",
        created_at: "2024-01-01",
        updated_at: "2024-01-01",
      },
    ];
    mockGet.mockResolvedValue({ success: true, data: { users } });
    const result = await listAllUsers();
    expect(result).toEqual(users);
    expect(mockGet).toHaveBeenCalledWith("/users");
  });

  it("returns empty array when data.users is absent", async () => {
    mockGet.mockResolvedValue({ success: true, data: {} });
    const result = await listAllUsers();
    expect(result).toEqual([]);
  });

  it("throws on failure", async () => {
    mockGet.mockResolvedValue({ success: false, error: "Unauthorized" });
    await expect(listAllUsers()).rejects.toThrow("Unauthorized");
  });

  it("throws with default message when no error provided", async () => {
    mockGet.mockResolvedValue({ success: false });
    await expect(listAllUsers()).rejects.toThrow("Failed to fetch users");
  });
});

describe("createUser", () => {
  beforeEach(() => vi.clearAllMocks());

  it("creates a user and returns it", async () => {
    const newUser = {
      id: "u2",
      tenant_id: "t2",
      username: "bob",
      created_at: "2024-01-01",
      updated_at: "2024-01-01",
    };
    mockPost.mockResolvedValue({ success: true, data: newUser });
    const result = await createUser({ username: "bob", password: "pass123" });
    expect(result).toEqual(newUser);
    expect(mockPost).toHaveBeenCalledWith("/users", {
      username: "bob",
      password: "pass123",
    });
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Already exists" });
    await expect(createUser({ username: "dup" })).rejects.toThrow(
      "Already exists",
    );
  });

  it("throws when data is null", async () => {
    mockPost.mockResolvedValue({ success: true, data: null });
    await expect(createUser({ username: "new" })).rejects.toThrow(
      "Failed to create user",
    );
  });
});

describe("updateUser", () => {
  beforeEach(() => vi.clearAllMocks());

  it("updates user and returns updated user", async () => {
    const updated = {
      id: "u3",
      tenant_id: "t3",
      email: "updated@example.com",
      created_at: "2024-01-01",
      updated_at: "2024-06-01",
    };
    mockPatch.mockResolvedValue({ success: true, data: updated });
    const result = await updateUser("t3", { email: "updated@example.com" });
    expect(result).toEqual(updated);
    expect(mockPatch).toHaveBeenCalledWith("/users/t3", {
      email: "updated@example.com",
    });
  });

  it("throws on failure", async () => {
    mockPatch.mockResolvedValue({ success: false, error: "Not found" });
    await expect(updateUser("t99", {})).rejects.toThrow("Not found");
  });

  it("throws when data is null", async () => {
    mockPatch.mockResolvedValue({ success: true, data: null });
    await expect(updateUser("t3", {})).rejects.toThrow("Failed to update user");
  });
});

describe("deleteUser", () => {
  beforeEach(() => vi.clearAllMocks());

  it("deletes user successfully (returns void)", async () => {
    mockDelete.mockResolvedValue({ success: true });
    const result = await deleteUser("t4");
    expect(result).toBeUndefined();
    expect(mockDelete).toHaveBeenCalledWith("/users/t4");
  });

  it("throws on failure", async () => {
    mockDelete.mockResolvedValue({ success: false, error: "Cannot delete" });
    await expect(deleteUser("t5")).rejects.toThrow("Cannot delete");
  });

  it("throws with default message when no error provided", async () => {
    mockDelete.mockResolvedValue({ success: false });
    await expect(deleteUser("t6")).rejects.toThrow("Failed to delete user");
  });
});

describe("changePassword", () => {
  beforeEach(() => vi.clearAllMocks());

  it("changes password successfully", async () => {
    mockPost.mockResolvedValue({ success: true });
    const result = await changePassword("tenant-01", "oldPass", "newPass");
    expect(result).toBeUndefined();
    expect(mockPost).toHaveBeenCalledWith("/users/tenant-01/change-password", {
      current_password: "oldPass",
      new_password: "newPass",
    });
  });

  it("does NOT include confirm_password in payload", async () => {
    mockPost.mockResolvedValue({ success: true });
    await changePassword("t7", "old", "new");
    const callArgs = mockPost.mock.calls[0];
    expect(callArgs[1]).not.toHaveProperty("confirm_password");
  });

  it("throws on failure", async () => {
    mockPost.mockResolvedValue({ success: false, error: "Wrong password" });
    await expect(changePassword("t8", "wrong", "new")).rejects.toThrow(
      "Wrong password",
    );
  });

  it("throws with default message when no error provided", async () => {
    mockPost.mockResolvedValue({ success: false });
    await expect(changePassword("t9", "x", "y")).rejects.toThrow(
      "Failed to change password",
    );
  });

  it("uses tenantId in URL path", async () => {
    mockPost.mockResolvedValue({ success: true });
    await changePassword("my-tenant-id", "p1", "p2");
    const url = mockPost.mock.calls[0][0];
    expect(url).toBe("/users/my-tenant-id/change-password");
  });
});
