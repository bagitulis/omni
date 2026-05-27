import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactElement } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CredentialPlatformSummary } from "@/api/credentials";
import { useAuthStore } from "@/stores/authStore";
import PlatformsTab from "./PlatformsTab";

type AuthStoreState = ReturnType<typeof useAuthStore.getState>;

function setAuthState(overrides: Partial<AuthStoreState> = {}) {
  useAuthStore.setState({
    user: { id: "1", username: "Dev", email: "dev@example.com", role: "developer" },
    token: "test-token",
    accessToken: "test-token",
    isAuthenticated: true,
    isInitializing: false,
    tenantId: "tenant-a",
    expiresAt: null,
    ...overrides,
  });
}

async function updateAuthStateAndRerender(
  rerender: (ui: ReactElement) => void,
  overrides: Partial<AuthStoreState>,
) {
  await act(async () => {
    useAuthStore.setState(overrides);
    rerender(<PlatformsTab />);
  });
}

const getCredentialPlatforms = vi.fn();
const getCredentialAudit = vi.fn();
const saveManualToken = vi.fn();

vi.mock("@/api/credentials", async () => {
  const actual = await vi.importActual<typeof import("@/api/credentials")>("@/api/credentials");
  return {
    ...actual,
    getCredentialPlatforms: (context: unknown) => getCredentialPlatforms(context),
    getCredentialAudit: (platform: string, context: unknown) => getCredentialAudit(platform, context),
    saveManualToken: (platform: string, payload: unknown, context: unknown) => saveManualToken(platform, payload, context),
  };
});

vi.mock("@/components/AntStaticApi", () => ({
  message: {
    error: vi.fn(),
    info: vi.fn(),
    success: vi.fn(),
  },
}));

const basePlatform = (overrides: Partial<CredentialPlatformSummary>): CredentialPlatformSummary => ({
  platform: "shopee",
  region: "id",
  status: "connected",
  app_configured: true,
  secret_mask: "configured",
  stores: [
    {
      store_identifier: "STORE-001",
      store_name: "[TEST DATA] Store",
      status: "connected",
      expires_at: "2026-06-01T00:00:00Z",
      last_refresh_at: "2026-05-26T00:00:00Z",
    },
  ],
  ...overrides,
});

const statusPlatforms: CredentialPlatformSummary[] = [
  basePlatform({ platform: "shopee", status: "disconnected", stores: [] }),
  basePlatform({ platform: "lazada", status: "incomplete", stores: [], secret_mask: "configured" }),
  basePlatform({ platform: "tiktok", status: "connected" }),
  basePlatform({ platform: "expired", status: "expired", stores: [] }),
  basePlatform({ platform: "refresh", status: "refresh_failed", stores: [] }),
  basePlatform({ platform: "action", status: "action_required", stores: [] }),
];

describe("PlatformsTab credential management", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    getCredentialAudit.mockResolvedValue([]);
    getCredentialPlatforms.mockResolvedValue(statusPlatforms);
    setAuthState();
  });

  it("separates store connections from app credentials and labels all credential states", async () => {
    render(<PlatformsTab />);

    expect(await screen.findByText("Store Connections")).toBeInTheDocument();
    expect(screen.getByText("App Credentials")).toBeInTheDocument();
    expect(screen.getAllByText("Disconnected").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Incomplete").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Connected").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Expired").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Refresh Failed").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Action Required").length).toBeGreaterThan(0);
  });

  it("shows Lazada as Indonesia while preserving masked secret display", async () => {
    render(<PlatformsTab />);

    await screen.findByText("Lazada App");
    expect(screen.getAllByText("Indonesia").length).toBeGreaterThan(0);
    expect(screen.getAllByText("configured").length).toBeGreaterThan(0);
    expect(screen.queryByText("raw-secret-value")).not.toBeInTheDocument();
  });

  it("disables manual token actions for tenant admins with deterministic reason text", async () => {
    setAuthState({
      user: { id: "2", username: "Admin", email: "admin@example.com", role: "owner" },
    });

    render(<PlatformsTab />);

    await screen.findByText("App Credentials");
    const manualButtons = screen.getAllByRole("button", { name: /manual token/i });
    expect(manualButtons[0]).toBeDisabled();
    expect(
      screen.getAllByText("Developer or system admin role required for app credentials and manual tokens.").length,
    ).toBeGreaterThan(0);
  });

  it("opens developer manual token drawer with write-only token fields", async () => {
    const user = userEvent.setup();
    render(<PlatformsTab />);

    const manualButtons = await screen.findAllByRole("button", { name: /manual token/i });
    await user.click(manualButtons[0]);

    expect(screen.getByText(/Manual Token - Shopee/i)).toBeInTheDocument();
    expect(screen.getByLabelText("Access Token")).toHaveAttribute("type", "password");
    expect(screen.getByLabelText("Refresh Token")).toHaveAttribute("type", "password");
    expect(screen.getByLabelText("Shop Cipher")).toHaveAttribute("type", "password");
  });

  it("loads OAuth and refresh history in the history drawer", async () => {
    getCredentialAudit.mockResolvedValueOnce([
      {
        event_type: "refresh_failed",
        platform: "shopee",
        store_identifier_mask: "***001",
        code: "token_expired",
        created_at: "2026-05-26T00:00:00Z",
      },
    ]);
    const user = userEvent.setup();

    render(<PlatformsTab />);
    const historyButtons = await screen.findAllByRole("button", { name: /^history$/i });
    await user.click(historyButtons[0]);

    await waitFor(() => expect(getCredentialAudit).toHaveBeenCalledWith("tiktok", { tenant_id: "tenant-a" }));
    expect(await screen.findByText("refresh_failed")).toBeInTheDocument();
    expect(screen.getByText("token_expired")).toBeInTheDocument();
  }, 15000);

  it("keys credential status refetch by current tenant and clears stale tenant data", async () => {
    let resolveTenantA: (value: CredentialPlatformSummary[]) => void = () => {};
    getCredentialPlatforms.mockReturnValueOnce(
      new Promise<CredentialPlatformSummary[]>((resolve) => {
        resolveTenantA = resolve;
      }),
    );
    getCredentialPlatforms.mockResolvedValueOnce([
      basePlatform({
        platform: "lazada",
        stores: [{ store_identifier: "STORE-B", store_name: "[TEST DATA] Tenant B", status: "connected" }],
      }),
    ]);

    const { rerender } = render(<PlatformsTab />);

    expect(getCredentialPlatforms).toHaveBeenCalledWith({ tenant_id: "tenant-a" });
    await updateAuthStateAndRerender(rerender, { tenantId: "tenant-b" });

    await waitFor(() =>
      expect(getCredentialPlatforms).toHaveBeenLastCalledWith({ tenant_id: "tenant-b" }),
    );
    expect(screen.queryByText("[TEST DATA] Store")).not.toBeInTheDocument();
    expect(await screen.findByText("[TEST DATA] Tenant B")).toBeInTheDocument();
    expect(screen.getByText("Credential status is scoped to tenant tenant-b.")).toBeInTheDocument();

    resolveTenantA([basePlatform({ stores: [{ store_identifier: "STORE-A", store_name: "[TEST DATA] Tenant A", status: "connected" }] })]);
    expect(screen.queryByText("[TEST DATA] Tenant A")).not.toBeInTheDocument();
  });

  it("allows developer and system admin manual token actions while tenant admin store actions stay available", async () => {
    const { rerender } = render(<PlatformsTab />);
    expect(await screen.findAllByRole("button", { name: /manual token/i })).toHaveLength(statusPlatforms.length);
    expect(screen.getAllByRole("button", { name: /connect/i })[0]).toBeEnabled();

    await updateAuthStateAndRerender(rerender, {
      user: { id: "3", username: "System Admin", email: "admin@example.com", role: "admin" },
    });

    await screen.findByText("App Credentials");
    expect(screen.getAllByRole("button", { name: /manual token/i })[0]).toBeEnabled();
  }, 15000);

  it("renders responsive columns for narrow layouts", async () => {
    window.innerWidth = 320;
    render(<PlatformsTab />);

    expect(await screen.findByText("Credential Management")).toBeInTheDocument();
    expect(screen.getByText("Store Connections")).toBeInTheDocument();
  });
});
