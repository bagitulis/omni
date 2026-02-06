import { ref, onMounted, type Ref } from "vue";
import api from "@/services/api";

export interface Sheet {
  sheetId: number;
  name: string;
  index: number;
}

export interface RegisteredSheet {
  id: string;
  url: string;
  spreadsheetId: string;
  spreadsheetName: string;
  sheets: Sheet[];
  purpose: string;
  createdBy: string;
  createdAt: Date;
  isLocked: boolean;
}

export interface RegistrationForm {
  url: string;
  purpose: string;
}

export interface DetectionResult {
  name: string;
  sheets: Sheet[];
}

export interface UseSpreadsheetRegistryReturn {
  showForm: Ref<boolean>;
  loading: Ref<boolean>;
  error: Ref<string>;
  success: Ref<string>;
  registeredSheets: Ref<RegisteredSheet[]>;
  editingId: Ref<string | null>;
  editingUrl: Ref<string>;
  registrationForm: Ref<RegistrationForm>;
  detectionResult: Ref<DetectionResult | null>;
  formatDate: (date: Date | string) => string;
  detectAndRegister: () => Promise<void>;
  loadSheets: () => Promise<void>;
  toggleEdit: (sheetId: string) => void;
  updateSheetUrl: (sheetId: string) => Promise<void>;
  deleteSheet: (sheetId: string) => Promise<void>;
}

export function useSpreadsheetRegistry(): UseSpreadsheetRegistryReturn {
  const showForm = ref(false);
  const loading = ref(false);
  const error = ref("");
  const success = ref("");
  const registeredSheets = ref<RegisteredSheet[]>([]);
  const editingId = ref<string | null>(null);
  const editingUrl = ref("");
  const registrationForm = ref<RegistrationForm>({ url: "", purpose: "" });
  const detectionResult = ref<DetectionResult | null>(null);

  function formatDate(date: Date | string): string {
    return new Date(date).toLocaleDateString();
  }

  async function detectAndRegister() {
    try {
      loading.value = true;
      error.value = "";
      success.value = "";

      const response = await api.post("/google/registry/register", {
        url: registrationForm.value.url,
        purpose: registrationForm.value.purpose,
      });

      if (response.data.success) {
        success.value = `✅ Sheet "${response.data.data.spreadsheetName}" registered successfully!`;
        registrationForm.value = { url: "", purpose: "" };
        detectionResult.value = null;
        await loadSheets();
        showForm.value = false;
      } else {
        error.value = response.data.error || "Failed to register sheet";
      }
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { error?: string } } };
      error.value = axiosErr.response?.data?.error || "Error registering sheet";
    } finally {
      loading.value = false;
    }
  }

  async function loadSheets() {
    try {
      const response = await api.get("/google/registry");
      if (response.data.success) {
        registeredSheets.value = response.data.data;
      }
    } catch (err) {
      console.error("Error loading sheets", err);
    }
  }

  function toggleEdit(sheetId: string) {
    const sheet = registeredSheets.value.find((s) => s.id === sheetId);
    if (sheet) {
      if (sheet.isLocked) {
        editingId.value = sheetId;
        editingUrl.value = sheet.url;
      } else {
        lockSheet(sheetId);
      }
    }
  }

  async function updateSheetUrl(sheetId: string) {
    try {
      const response = await api.post(`/google/registry/${sheetId}/unlock`, {
        url: editingUrl.value,
      });
      if (response.data.success) {
        editingId.value = null;
        await loadSheets();
      }
    } catch {
      error.value = "Failed to update sheet";
    }
  }

  async function lockSheet(sheetId: string) {
    try {
      await api.post(`/google/registry/${sheetId}/lock`);
      await loadSheets();
    } catch {
      error.value = "Failed to lock sheet";
    }
  }

  async function deleteSheet(sheetId: string) {
    if (confirm("Are you sure you want to delete this sheet registration?")) {
      try {
        await api.delete(`/google/registry/${sheetId}`);
        await loadSheets();
        success.value = "Sheet deleted";
      } catch {
        error.value = "Failed to delete sheet";
      }
    }
  }

  onMounted(() => {
    loadSheets();
  });

  return {
    showForm,
    loading,
    error,
    success,
    registeredSheets,
    editingId,
    editingUrl,
    registrationForm,
    detectionResult,
    formatDate,
    detectAndRegister,
    loadSheets,
    toggleEdit,
    updateSheetUrl,
    deleteSheet,
  };
}
