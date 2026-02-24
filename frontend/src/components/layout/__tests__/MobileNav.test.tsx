import { render, screen } from "@testing-library/react";
import { theme } from "antd";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import MobileNav from "../MobileNav";
import {
  formatTimeRemaining,
  getStatusColor,
  getStatusIcon,
  getTokenStatusType,
} from "../TokenStatusDropdown.utils";
import type { TokenStatusData } from "../TokenStatusDropdown.types";

const navigateMock = vi.fn();
const BASE_TIME = new Date("2026-01-01T00:00:00.000Z");

function createTokenStatus(partial: Partial<TokenStatusData>): TokenStatusData {
  return {
    isExpired: false,
    expiresAt: null,
    refreshTokenExpiresAt: null,
    status: "valid",
    valid: true,
    ...partial,
  };
}

vi.mock("react-router-dom", () => ({
  useNavigate: () => navigateMock,
  useLocation: () => ({ pathname: "/script-monitor" }),
}));

describe("MobileNav", () => {
  it("renders Script Monitor link for mobile navigation", () => {
    render(<MobileNav />);

    expect(screen.getByText("Scripts")).toBeTruthy();
  });
});

describe("TokenStatusDropdown utils", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(BASE_TIME);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("resolves token status type across unknown, configured, and expiry states", () => {
    expect(getTokenStatusType(undefined)).toBe("unknown");
    expect(
      getTokenStatusType(createTokenStatus({ status: "not_configured" })),
    ).toBe("not_configured");
    expect(getTokenStatusType(createTokenStatus({ isExpired: true }))).toBe(
      "expired",
    );
    expect(getTokenStatusType(createTokenStatus({ valid: false }))).toBe(
      "expired",
    );

    const in6Hours = new Date(BASE_TIME.getTime() + 6 * 60 * 60 * 1000);
    expect(
      getTokenStatusType(
        createTokenStatus({ expiresAt: in6Hours.toISOString() }),
      ),
    ).toBe("expiring");

    const in2Days = new Date(BASE_TIME.getTime() + 2 * 24 * 60 * 60 * 1000);
    expect(
      getTokenStatusType(
        createTokenStatus({ expiresAt: in2Days.toISOString() }),
      ),
    ).toBe("valid");

    const expired1MinuteAgo = new Date(BASE_TIME.getTime() - 60 * 1000);
    expect(
      getTokenStatusType(
        createTokenStatus({ expiresAt: expired1MinuteAgo.toISOString() }),
      ),
    ).toBe("expired");
  });

  it("formats remaining time for null, expired, days, hours, and minutes branches", () => {
    expect(formatTimeRemaining(null)).toBe("Unknown");

    const expired = new Date(BASE_TIME.getTime() - 1000);
    expect(formatTimeRemaining(expired.toISOString())).toBe("Expired");

    const in2Days3Hours = new Date(
      BASE_TIME.getTime() + (2 * 24 + 3) * 60 * 60 * 1000,
    );
    expect(formatTimeRemaining(in2Days3Hours.toISOString())).toBe("2d 3h");

    const in5Hours15Minutes = new Date(
      BASE_TIME.getTime() + (5 * 60 + 15) * 60 * 1000,
    );
    expect(formatTimeRemaining(in5Hours15Minutes.toISOString())).toBe("5h 15m");

    const in45Minutes = new Date(BASE_TIME.getTime() + 45 * 60 * 1000);
    expect(formatTimeRemaining(in45Minutes.toISOString())).toBe("45m");
  });

  it("maps status colors and icon colors correctly", () => {
    expect(getStatusColor("valid")).toBe("success");
    expect(getStatusColor("expiring")).toBe("warning");
    expect(getStatusColor("expired")).toBe("error");
    expect(getStatusColor("not_configured")).toBe("default");
    expect(getStatusColor("unknown")).toBe("default");

    const token = theme.getDesignToken();

    expect(getStatusIcon("valid", token).props.style.color).toBe(
      token.colorSuccess,
    );
    expect(getStatusIcon("expiring", token).props.style.color).toBe(
      token.colorWarning,
    );
    expect(getStatusIcon("expired", token).props.style.color).toBe(
      token.colorError,
    );
    expect(getStatusIcon("not_configured", token).props.style.color).toBe(
      token.colorBorder,
    );
    expect(getStatusIcon("other", token).props.style.color).toBe(
      token.colorTextSecondary,
    );
  });
});
