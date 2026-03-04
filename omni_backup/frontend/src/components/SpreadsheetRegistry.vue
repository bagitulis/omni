<template>
  <div class="spreadsheet-registry">
    <div v-if="!showForm" class="registry-header">
      <h2>📊 Sheet Configuration</h2>
      <button @click="showForm = true" class="btn-primary">
        + Register New Sheet
      </button>
    </div>

    <div v-else class="register-form">
      <h3>Register Google Sheet</h3>
      <div class="form-group">
        <label for="sheet-url-input">Sheet URL</label>
        <input
          id="sheet-url-input"
          v-model="registrationForm.url"
          type="text"
          placeholder="https://docs.google.com/spreadsheets/d/..."
          class="form-input"
        />
        <small>Paste the full Google Sheets URL</small>
      </div>

      <div class="form-group">
        <label for="sheet-purpose-select">Purpose</label>
        <select
          id="sheet-purpose-select"
          v-model="registrationForm.purpose"
          class="form-input"
        >
          <option value="">Select purpose...</option>
          <option value="inventory">Inventory</option>
          <option value="wallet">Wallet</option>
          <option value="shipping">Shipping</option>
          <option value="order">Order</option>
          <option value="custom">Custom</option>
        </select>
      </div>

      <div v-if="detectionResult" class="detection-result">
        <div class="result-item">
          <strong>✅ Spreadsheet:</strong> {{ detectionResult.name }}
        </div>
        <div class="result-item">
          <strong>📄 Sheets:</strong>
          {{ detectionResult.sheets.map((s) => s.name).join(", ") }}
        </div>
      </div>

      <div class="form-actions">
        <button
          @click="detectAndRegister"
          :disabled="
            !registrationForm.url || !registrationForm.purpose || loading
          "
          class="btn-primary"
        >
          {{ loading ? "⏳ Detecting..." : "✨ Detect & Register" }}
        </button>
        <button @click="showForm = false" class="btn-secondary">Cancel</button>
      </div>

      <div v-if="error" class="alert-error">{{ error }}</div>
      <div v-if="success" class="alert-success">{{ success }}</div>
    </div>

    <div v-if="registeredSheets.length > 0" class="sheets-list">
      <h3>Registered Sheets</h3>
      <div v-for="sheet in registeredSheets" :key="sheet.id" class="sheet-card">
        <div class="sheet-header">
          <div class="sheet-info">
            <h4>{{ sheet.spreadsheetName }}</h4>
            <p class="sheet-meta">
              Purpose: <span class="badge">{{ sheet.purpose }}</span> | Created:
              {{ formatDate(sheet.createdAt) }}
            </p>
          </div>
          <div class="sheet-actions">
            <button
              @click="toggleEdit(sheet.id)"
              class="btn-small"
              :class="{
                'btn-unlock': sheet.isLocked,
                'btn-lock': !sheet.isLocked,
              }"
            >
              {{ sheet.isLocked ? "🔒 Locked" : "🔓 Unlocked" }}
            </button>
            <button @click="deleteSheet(sheet.id)" class="btn-small btn-danger">
              🗑️ Delete
            </button>
          </div>
        </div>

        <div class="sheet-details">
          <strong>Sheets in this spreadsheet:</strong>
          <div class="sheets-grid">
            <div v-for="s in sheet.sheets" :key="s.sheetId" class="sheet-item">
              {{ s.name }}
            </div>
          </div>
        </div>

        <div v-if="editingId === sheet.id" class="edit-url">
          <input
            v-model="editingUrl"
            type="text"
            class="form-input"
            placeholder="New URL"
          />
          <button
            @click="updateSheetUrl(sheet.id)"
            class="btn-small btn-primary"
          >
            💾 Save
          </button>
          <button @click="editingId = null" class="btn-small btn-secondary">
            Cancel
          </button>
        </div>
      </div>
    </div>

    <div v-else-if="!showForm" class="empty-state">
      <p>No sheets registered yet</p>
      <button @click="showForm = true" class="btn-primary">
        Register Your First Sheet
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSpreadsheetRegistry } from "./composables/useSpreadsheetRegistry";

const {
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
  toggleEdit,
  updateSheetUrl,
  deleteSheet,
} = useSpreadsheetRegistry();
</script>

<style src="./SpreadsheetRegistry.styles.css" scoped></style>
