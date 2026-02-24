import { describe, it, expect, vi, beforeEach } from "vitest";
import { useModalsStore } from "./modalsStore";

const resetStore = () => {
  useModalsStore.setState({ modals: {} });
};

describe("modalsStore — initial state", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore();
  });

  it("modals starts as empty object", () => {
    expect(useModalsStore.getState().modals).toEqual({});
  });
});

describe("modalsStore — openModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore();
  });

  it("opens a modal by name", () => {
    useModalsStore.getState().openModal("testModal");
    expect(useModalsStore.getState().modals["testModal"].isOpen).toBe(true);
  });

  it("stores data when opening a modal", () => {
    useModalsStore.getState().openModal("editModal", { id: 42 });
    expect(useModalsStore.getState().modals["editModal"].data).toEqual({
      id: 42,
    });
  });

  it("single-modal enforcement: closes other modals when opening new one", () => {
    useModalsStore.getState().openModal("modalA");
    useModalsStore.getState().openModal("modalB");

    const state = useModalsStore.getState();
    expect(state.modals["modalA"].isOpen).toBe(false);
    expect(state.modals["modalB"].isOpen).toBe(true);
  });

  it("single-modal enforcement: multiple modals open then last one wins", () => {
    useModalsStore.getState().openModal("modal1");
    useModalsStore.getState().openModal("modal2");
    useModalsStore.getState().openModal("modal3");

    const state = useModalsStore.getState();
    expect(state.modals["modal1"].isOpen).toBe(false);
    expect(state.modals["modal2"].isOpen).toBe(false);
    expect(state.modals["modal3"].isOpen).toBe(true);
  });

  it("can re-open a previously closed modal", () => {
    useModalsStore.getState().openModal("modal1");
    useModalsStore.getState().closeModal("modal1");
    useModalsStore.getState().openModal("modal1");
    expect(useModalsStore.getState().modals["modal1"].isOpen).toBe(true);
  });

  it("stores undefined data when no data provided", () => {
    useModalsStore.getState().openModal("noDataModal");
    expect(
      useModalsStore.getState().modals["noDataModal"].data,
    ).toBeUndefined();
  });
});

describe("modalsStore — closeModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore();
  });

  it("closes an open modal", () => {
    useModalsStore.getState().openModal("closeTest");
    useModalsStore.getState().closeModal("closeTest");
    expect(useModalsStore.getState().modals["closeTest"].isOpen).toBe(false);
  });

  it("preserves modal data after closing", () => {
    useModalsStore.getState().openModal("dataModal", { value: "test" });
    useModalsStore.getState().closeModal("dataModal");
    expect(useModalsStore.getState().modals["dataModal"].data).toEqual({
      value: "test",
    });
  });

  it("does not affect other modals when closing one", () => {
    // Open two modals sequentially (so B closes A)
    useModalsStore.getState().openModal("modalX");
    // Manually set both open to test isolation
    useModalsStore.setState({
      modals: {
        modalX: { isOpen: true },
        modalY: { isOpen: true },
      },
    });
    useModalsStore.getState().closeModal("modalX");
    expect(useModalsStore.getState().modals["modalY"].isOpen).toBe(true);
  });

  it("closing non-existent modal sets it to isOpen false", () => {
    useModalsStore.getState().closeModal("ghost");
    expect(useModalsStore.getState().modals["ghost"].isOpen).toBe(false);
  });
});

describe("modalsStore — getModalData", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore();
  });

  it("returns data for open modal", () => {
    useModalsStore.getState().openModal("withData", { foo: "bar" });
    expect(useModalsStore.getState().getModalData("withData")).toEqual({
      foo: "bar",
    });
  });

  it("returns undefined for modal that was never opened", () => {
    expect(useModalsStore.getState().getModalData("never")).toBeUndefined();
  });

  it("returns undefined when modal opened without data", () => {
    useModalsStore.getState().openModal("noData");
    expect(useModalsStore.getState().getModalData("noData")).toBeUndefined();
  });

  it("returns complex data objects", () => {
    const complexData = { user: { id: 1, name: "Alice" }, items: [1, 2, 3] };
    useModalsStore.getState().openModal("complexModal", complexData);
    expect(useModalsStore.getState().getModalData("complexModal")).toEqual(
      complexData,
    );
  });
});

describe("modalsStore — isModalOpen", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    resetStore();
  });

  it("returns true for an open modal", () => {
    useModalsStore.getState().openModal("openModal");
    expect(useModalsStore.getState().isModalOpen("openModal")).toBe(true);
  });

  it("returns false for a closed modal", () => {
    useModalsStore.getState().openModal("closeMe");
    useModalsStore.getState().closeModal("closeMe");
    expect(useModalsStore.getState().isModalOpen("closeMe")).toBe(false);
  });

  it("returns false for a modal that was never opened", () => {
    expect(useModalsStore.getState().isModalOpen("doesNotExist")).toBe(false);
  });

  it("returns false when single-modal enforcement closes it", () => {
    useModalsStore.getState().openModal("first");
    useModalsStore.getState().openModal("second");
    expect(useModalsStore.getState().isModalOpen("first")).toBe(false);
    expect(useModalsStore.getState().isModalOpen("second")).toBe(true);
  });
});
