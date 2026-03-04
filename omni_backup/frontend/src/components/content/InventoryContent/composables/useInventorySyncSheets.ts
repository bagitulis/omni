import { getAuthHeaders, getApiBaseUrl } from "../../../../utils/apiHeaders";

export function useInventorySyncSheets(
  state: any,
  alertHandlers: any,
  dataMethods: any,
  configMethods?: any, // Optional: for getting locked columns
) {
  const API_BASE_URL = getApiBaseUrl("/inventory");

  async function handleSyncFromSheets() {
    try {
      if (!state.configData.spreadsheet_id || !state.configData.sheet_name) {
        alertHandlers.error(
          "Konfigurasi Belum Lengkap",
          "Silakan setup spreadsheet di Settings page terlebih dahulu (⚙️ Pengaturan)",
        );
        return;
      }

      state.syncing = true;
      alertHandlers.info("Syncing...", "Loading data from Google Sheets...");

      const response = await fetch(`${API_BASE_URL}/sync/from-sheets`, {
        method: "POST",
        headers: {
          ...getAuthHeaders(),
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          spreadsheet_id: state.configData.spreadsheet_id,
          sheet_name: state.configData.sheet_name,
          header_row: state.configData.header_row || 1,
          start_row: state.configData.data_start_row || 2,
        }),
      });

      const result = await response.json();

      // Handle success, partial success, and error statuses
      const isSuccess =
        result.status === "success" || result.status === "partial";

      if (isSuccess) {
        const failedMsg =
          result.failed_records > 0
            ? `\n⚠️ ${result.failed_records} rows skipped (possibly empty rows or incomplete data - this is normal)`
            : "";

        // Use warning for partial, success for full success
        const alertType = result.status === "partial" ? "warning" : "success";
        const title =
          result.status === "partial"
            ? "⚠️ Sync Partial"
            : "✅ Sync Successful!";

        if (alertType === "warning") {
          alertHandlers.warning(
            title,
            `${result.synced_records} records synchronized from Google Sheets${failedMsg}`,
            {
              synced_records: result.synced_records,
              failed_records: result.failed_records,
              duration: result.duration,
            },
          );
        } else {
          alertHandlers.success(
            title,
            `${result.synced_records} records successfully synchronized from Google Sheets${failedMsg}`,
            {
              synced_records: result.synced_records,
              failed_records: result.failed_records,
              duration: result.duration,
            },
          );
        }

        // Update syncStatus - treat partial as success for UI purposes
        state.syncStatus = {
          status: result.status === "partial" ? "partial" : "success",
          timestamp: new Date().toISOString(),
          message: `${result.synced_records} records synced`,
        };

        // Reload data from database
        await dataMethods.loadInventoryData();
        await dataMethods.loadStats();

        if (result.headers_changed) {
          alertHandlers.warning(
            "Schema Changed!",
            "Database schema telah diperbarui otomatis sesuai kolom baru",
          );
        }
      } else {
        // Handle actual errors
        state.syncStatus = {
          status: "error",
          timestamp: new Date().toISOString(),
          message: result.message,
        };
        alertHandlers.error("Sync Gagal", result.message);
      }
    } catch (error: any) {
      console.error("❌ Sync error:", error);
      state.syncStatus = {
        status: "error",
        timestamp: new Date().toISOString(),
        message: error.message,
      };
      alertHandlers.error("Sync Error", error.message);
    } finally {
      state.syncing = false;
    }
  }

  async function handleSyncToSheets() {
    try {
      if (!state.configData.spreadsheet_id || !state.configData.sheet_name) {
        alertHandlers.error(
          "Konfigurasi Belum Lengkap",
          "Silakan setup spreadsheet di Settings page terlebih dahulu (⚙️ Pengaturan)",
        );
        return;
      }

      state.syncing = true;
      alertHandlers.info("Exporting...", "Delta sync to Google Sheets...");

      // Get locked columns from config (if available)
      const lockedColumns = configMethods?.getLockedColumns?.() || [];

      const response = await fetch(`${API_BASE_URL}/sync/to-sheets`, {
        method: "POST",
        headers: {
          ...getAuthHeaders(),
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          spreadsheet_id: state.configData.spreadsheet_id,
          sheet_name: state.configData.sheet_name,
          locked_columns: lockedColumns,
        }),
      });

      const result = await response.json();

      if (result.status === "success") {
        // Enhanced message with delta sync details
        const detailMsg = `${result.updated_records || 0} cells updated, ${result.new_records || 0} new rows`;
        alertHandlers.success(
          "Export Berhasil! ✅",
          `${result.message}\n${detailMsg}`,
          {
            synced_records: result.synced_records,
            updated_records: result.updated_records,
            new_records: result.new_records,
            unchanged_records: result.unchanged_records,
          },
        );

        // Update syncStatus for export
        state.syncStatus = {
          status: "success",
          timestamp: new Date().toISOString(),
          message: `${result.synced_records} records exported`,
        };

        // Reload stats after export
        await dataMethods.loadStats();
      } else {
        state.syncStatus = {
          status: "error",
          timestamp: new Date().toISOString(),
          message: result.message,
        };
        alertHandlers.error("Export Gagal", result.message);
      }
    } catch (error: any) {
      console.error("❌ Export error:", error);
      state.syncStatus = {
        status: "error",
        timestamp: new Date().toISOString(),
        message: error.message,
      };
      alertHandlers.error("Export Error", error.message);
    } finally {
      state.syncing = false;
    }
  }

  return {
    handleSyncFromSheets,
    handleSyncToSheets,
  };
}
