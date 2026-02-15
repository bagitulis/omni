import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MobileNav from "../MobileNav";

const navigateMock = vi.fn();

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
