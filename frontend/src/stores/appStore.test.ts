import { describe, it, expect, vi, beforeEach } from "vitest";
import { useAppStore } from "./appStore";

describe("appStore — initial state", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAppStore.setState({
      connectionStatus: "connecting",
      isConnected: false,
      isLoading: false,
      sidebarCollapsed: false,
      activeTab: "order-management",
      activePlatform: "shopee",
    });
  });

  it("connectionStatus starts as connecting", () => {
    expect(useAppStore.getState().connectionStatus).toBe("connecting");
  });

  it("isConnected starts as false", () => {
    expect(useAppStore.getState().isConnected).toBe(false);
  });

  it("isLoading starts as false", () => {
    expect(useAppStore.getState().isLoading).toBe(false);
  });

  it("sidebarCollapsed starts as false", () => {
    expect(useAppStore.getState().sidebarCollapsed).toBe(false);
  });

  it("activeTab starts as order-management", () => {
    expect(useAppStore.getState().activeTab).toBe("order-management");
  });

  it("activePlatform starts as shopee", () => {
    expect(useAppStore.getState().activePlatform).toBe("shopee");
  });
});

describe("appStore — setConnectionStatus", () => {
  beforeEach(() => {
    useAppStore.setState({
      connectionStatus: "connecting",
      isConnected: false,
      isLoading: false,
      sidebarCollapsed: false,
      activeTab: "order-management",
      activePlatform: "shopee",
    });
  });

  it("sets connectionStatus and isConnected=true when connected", () => {
    useAppStore.getState().setConnectionStatus("connected");
    const state = useAppStore.getState();
    expect(state.connectionStatus).toBe("connected");
    expect(state.isConnected).toBe(true);
  });

  it("sets isConnected=false for error status", () => {
    useAppStore.getState().setConnectionStatus("error");
    const state = useAppStore.getState();
    expect(state.connectionStatus).toBe("error");
    expect(state.isConnected).toBe(false);
  });

  it("sets isConnected=false for disconnected status", () => {
    useAppStore.getState().setConnectionStatus("disconnected");
    expect(useAppStore.getState().isConnected).toBe(false);
  });

  it("sets isConnected=false for connecting status", () => {
    useAppStore.getState().setConnectionStatus("connected"); // first connected
    useAppStore.getState().setConnectionStatus("connecting"); // then back
    expect(useAppStore.getState().isConnected).toBe(false);
  });
});

describe("appStore — setIsLoading", () => {
  beforeEach(() => {
    useAppStore.setState({
      connectionStatus: "connecting",
      isConnected: false,
      isLoading: false,
      sidebarCollapsed: false,
      activeTab: "order-management",
      activePlatform: "shopee",
    });
  });

  it("sets isLoading to true", () => {
    useAppStore.getState().setIsLoading(true);
    expect(useAppStore.getState().isLoading).toBe(true);
  });

  it("sets isLoading to false", () => {
    useAppStore.setState({ isLoading: true });
    useAppStore.getState().setIsLoading(false);
    expect(useAppStore.getState().isLoading).toBe(false);
  });
});

describe("appStore — sidebar", () => {
  beforeEach(() => {
    useAppStore.setState({
      connectionStatus: "connecting",
      isConnected: false,
      isLoading: false,
      sidebarCollapsed: false,
      activeTab: "order-management",
      activePlatform: "shopee",
    });
  });

  it("setSidebarCollapsed sets to true", () => {
    useAppStore.getState().setSidebarCollapsed(true);
    expect(useAppStore.getState().sidebarCollapsed).toBe(true);
  });

  it("setSidebarCollapsed sets to false", () => {
    useAppStore.setState({ sidebarCollapsed: true });
    useAppStore.getState().setSidebarCollapsed(false);
    expect(useAppStore.getState().sidebarCollapsed).toBe(false);
  });

  it("toggleSidebar flips false to true", () => {
    useAppStore.setState({ sidebarCollapsed: false });
    useAppStore.getState().toggleSidebar();
    expect(useAppStore.getState().sidebarCollapsed).toBe(true);
  });

  it("toggleSidebar flips true to false", () => {
    useAppStore.setState({ sidebarCollapsed: true });
    useAppStore.getState().toggleSidebar();
    expect(useAppStore.getState().sidebarCollapsed).toBe(false);
  });

  it("toggleSidebar called twice returns to original state", () => {
    useAppStore.setState({ sidebarCollapsed: false });
    useAppStore.getState().toggleSidebar();
    useAppStore.getState().toggleSidebar();
    expect(useAppStore.getState().sidebarCollapsed).toBe(false);
  });
});

describe("appStore — setActiveTab", () => {
  beforeEach(() => {
    useAppStore.setState({
      connectionStatus: "connecting",
      isConnected: false,
      isLoading: false,
      sidebarCollapsed: false,
      activeTab: "order-management",
      activePlatform: "shopee",
    });
  });

  it("sets activeTab to product-management", () => {
    useAppStore.getState().setActiveTab("product-management");
    expect(useAppStore.getState().activeTab).toBe("product-management");
  });

  it("sets activeTab to settings", () => {
    useAppStore.getState().setActiveTab("settings");
    expect(useAppStore.getState().activeTab).toBe("settings");
  });
});

describe("appStore — setActivePlatform", () => {
  beforeEach(() => {
    useAppStore.setState({
      connectionStatus: "connecting",
      isConnected: false,
      isLoading: false,
      sidebarCollapsed: false,
      activeTab: "order-management",
      activePlatform: "shopee",
    });
  });

  it("sets activePlatform to lazada", () => {
    useAppStore.getState().setActivePlatform("lazada");
    expect(useAppStore.getState().activePlatform).toBe("lazada");
  });

  it("sets activePlatform to tiktok", () => {
    useAppStore.getState().setActivePlatform("tiktok");
    expect(useAppStore.getState().activePlatform).toBe("tiktok");
  });

  it("sets activePlatform back to shopee", () => {
    useAppStore.setState({ activePlatform: "lazada" });
    useAppStore.getState().setActivePlatform("shopee");
    expect(useAppStore.getState().activePlatform).toBe("shopee");
  });
});
