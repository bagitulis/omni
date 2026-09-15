import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PlatformCard, type PlatformConnectionSummary } from "./PlatformCard";

// Phase 10.2 — render-level tests for the token-expired banner + CTA relabel.
// Backend `CredentialConnection.EffectiveStatus()` can emit "refresh_required"
// (access dead, refresh still alive) or "expired" (refresh_token dead — needs
// full re-OAuth). The card must surface both cases with a clear banner and a
// primary CTA labelled "Re-authorize" instead of the generic "Connect".

function summary(overrides: Partial<PlatformConnectionSummary>): PlatformConnectionSummary {
  return {
    platform: "shopee",
    connected: false,
    status: "disconnected",
    app_configured: true,
    ...overrides,
  };
}

const noop = () => undefined;
const formatExpiry = () => null;

describe("PlatformCard token-expired banner", () => {
  it("renders '🔴 Access token expired' banner when status is 'expired'", () => {
    render(
      <PlatformCard
        platform={summary({ status: "expired" })}
        color="#f43f5e"
        icon={<span data-testid="icon" />}
        name="Shopee"
        onConnect={noop}
        onDisconnect={noop}
        formatExpiry={formatExpiry}
        onViewHistory={noop}
      />,
    );
    // Banner should mention "Access token expired" and be visible.
    expect(screen.getByText(/Access token expired/i)).toBeInTheDocument();
    // Should reference the re-authorize action so the user knows what to do
    // (narrow to the button role — the banner and button both mention it).
    expect(
      screen.getByRole("button", { name: /Re-authorize Shopee/i }),
    ).toBeInTheDocument();
  });

  it("renders a softer 'Access token expired' banner when status is 'refresh_required'", () => {
    render(
      <PlatformCard
        platform={summary({ status: "refresh_required" })}
        color="#000"
        icon={<span data-testid="icon" />}
        name="TikTok"
        onConnect={noop}
        onDisconnect={noop}
        formatExpiry={formatExpiry}
        onViewHistory={noop}
      />,
    );
    // Same banner text; both statuses mean "user should re-authorize now".
    expect(screen.getByText(/Access token expired/i)).toBeInTheDocument();
  });

  it("uses 'Re-authorize' as primary CTA label when status is 'expired' and app_configured", () => {
    render(
      <PlatformCard
        platform={summary({ status: "expired", app_configured: true })}
        color="#f43f5e"
        icon={<span data-testid="icon" />}
        name="Shopee"
        onConnect={noop}
        onDisconnect={noop}
        formatExpiry={formatExpiry}
        onViewHistory={noop}
      />,
    );
    // A primary "Re-authorize Shopee" button should exist (not generic "Connect Shopee").
    expect(
      screen.getByRole("button", { name: /Re-authorize Shopee/i }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^Connect Shopee$/i }),
    ).not.toBeInTheDocument();
  });

  it("keeps the generic 'Connect Shopee' CTA when status is 'disconnected'", () => {
    render(
      <PlatformCard
        platform={summary({ status: "disconnected" })}
        color="#f43f5e"
        icon={<span data-testid="icon" />}
        name="Shopee"
        onConnect={noop}
        onDisconnect={noop}
        formatExpiry={formatExpiry}
        onViewHistory={noop}
      />,
    );
    expect(
      screen.getByRole("button", { name: /Connect Shopee/i }),
    ).toBeInTheDocument();
  });

  it("triggers onConnect when the Re-authorize CTA is clicked", async () => {
    const onConnect = vi.fn();
    render(
      <PlatformCard
        platform={summary({ status: "expired" })}
        color="#f43f5e"
        icon={<span data-testid="icon" />}
        name="Shopee"
        onConnect={onConnect}
        onDisconnect={noop}
        formatExpiry={formatExpiry}
        onViewHistory={noop}
      />,
    );
    const btn = screen.getByRole("button", { name: /Re-authorize Shopee/i });
    btn.click();
    expect(onConnect).toHaveBeenCalledTimes(1);
  });
});
