<template>
  <div class="section">
    <div class="form-group">
      <label :for="`${type}-link`">{{ label }}</label>
      <div class="link-input-group">
        <input
          :id="`${type}-link`"
          :value="modelValue"
          type="text"
          :placeholder="placeholder"
          :disabled="locked && validated"
          class="form-input"
          @input="
            $emit(
              'update:modelValue',
              ($event.target as HTMLInputElement).value
            )
          "
        />
        <!-- Button Group when validated and unlocked (show both Select & Lock) -->
        <div v-if="validated && !locked && modelValue" class="button-group">
          <button
            @click="$emit('lock')"
            class="btn-lock"
            :title="`Lock ${type}`"
          >
            🔒
          </button>
        </div>
        <!-- Unlock button when locked -->
        <button
          v-else-if="locked && validated && modelValue"
          @click="$emit('unlock')"
          class="btn-unlock"
          :title="`Edit ${type}`"
        >
          ✏️
        </button>
        <!-- Validate button when not validated -->
        <button
          v-else-if="modelValue && !validated"
          @click="$emit('check')"
          :disabled="validating"
          class="btn-check"
          :title="`Validate ${type}`"
        >
          {{ validating ? "⏳" : "✓" }}
        </button>
      </div>
      <p class="helper-text">{{ helperText }}</p>
    </div>

    <div v-if="validated" class="validation-result success">
      <p class="success">✅ {{ metadata?.name }}</p>
      <div v-if="metadata?.sheets && type === 'inventory'" class="sheets-list">
        <label for="sheet-select" class="sheets-label">Available sheets:</label>
        <select
          id="sheet-select"
          :value="selectedSheet"
          :disabled="locked"
          class="sheet-select"
          @change="handleSheetSelect"
        >
          <option value="">-- Select Sheet --</option>
          <option
            v-for="sheet in metadata.sheets"
            :key="sheet.sheetId"
            :value="sheet.name"
          >
            {{ sheet.name }} ({{ sheet.rowCount }} rows)
          </option>
        </select>
      </div>
    </div>

    <div v-if="error" class="validation-result error">
      <p class="error-msg">❌ {{ error }}</p>
    </div>
  </div>
</template>

<script lang="ts">
interface SheetMetadata {
  name: string;
  sheetId: number;
  index: number;
  columnCount: number;
  rowCount: number;
}

interface ValidationResult {
  spreadsheetId: string;
  name: string;
  sheets: SheetMetadata[];
  type: string;
}

export default {
  name: "SpreadsheetLinkField",
  props: {
    type: { type: String, required: true },
    label: { type: String, required: true },
    placeholder: {
      type: String,
      default: "https://docs.google.com/spreadsheets/d/...",
    },
    helperText: { type: String, required: true },
    modelValue: { type: String, default: "" },
    validating: { type: Boolean, default: false },
    validated: { type: Boolean, default: false },
    error: { type: String, default: "" },
    metadata: { type: Object as () => ValidationResult | null, default: null },
    selectedSheet: { type: String, default: "" },
    locked: { type: Boolean, default: false },
  },
  emits: ["update:modelValue", "check", "lock", "select-sheet", "unlock"],
  methods: {
    handleSheetSelect(event: Event) {
      const value = (event.target as HTMLSelectElement).value;
      this.$emit("select-sheet", value);
    },
  },
};
</script>

<style scoped>
.section {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.form-group {
  margin-bottom: 0;
}

label {
  display: block;
  font-weight: 600;
  margin-bottom: 8px;
  color: #333;
  font-size: 14px;
}

.form-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
}

.form-input:disabled {
  background-color: #f5f5f5;
  color: #6b7280;
  cursor: not-allowed;
  border-color: #ccc;
}

.form-input:focus {
  outline: none;
  border-color: #4285f4;
  box-shadow: 0 0 0 2px rgba(66, 133, 244, 0.1);
}

.form-input:disabled:focus {
  box-shadow: none;
  border-color: #ccc;
}

.link-input-group {
  display: flex;
  gap: 8px;
}

.link-input-group .form-input {
  flex: 1;
}

.button-group {
  display: flex;
  gap: 6px;
}

.btn-check {
  padding: 8px 12px;
  background: #34a853;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  min-width: 50px;
  transition: all 0.2s;
}

.btn-check:hover:not(:disabled) {
  background: #2d8e47;
  box-shadow: 0 2px 6px rgba(52, 168, 83, 0.3);
}

.btn-check:disabled {
  background: #ccc;
  cursor: not-allowed;
  opacity: 0.6;
}

.btn-unlock {
  padding: 8px 12px;
  background: #ff9800;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  min-width: 50px;
  transition: all 0.2s;
}

.btn-unlock:hover {
  background: #f57c00;
  box-shadow: 0 2px 6px rgba(255, 152, 0, 0.3);
}

.btn-lock {
  padding: 8px 12px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 600;
  min-width: 50px;
  transition: all 0.2s;
}

.btn-lock:hover {
  background: #1565c0;
  box-shadow: 0 2px 6px rgba(25, 118, 210, 0.3);
}

.helper-text {
  font-size: 12px;
  color: #6b7280;
  margin: 6px 0 0;
}

.validation-result {
  margin-top: 8px;
  padding: 10px 12px;
  border-radius: 6px;
  font-size: 13px;
}

.validation-result.success {
  background: #e8f5e9;
  border: 1px solid #c8e6c9;
}

.validation-result.error {
  background: #ffebee;
  border: 1px solid #ffcdd2;
}

.validation-result p {
  margin: 0 0 8px;
  font-weight: 500;
}

.validation-result p.success {
  color: #2e7d32;
}

.validation-result p.error-msg {
  color: #c62828;
  margin: 0;
}

.sheets-list {
  margin-top: 8px;
}

.sheets-label {
  font-weight: 500;
  color: #333;
  margin-bottom: 6px;
  font-size: 12px;
  margin: 0;
}

.sheet-select {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 13px;
  background: white;
  cursor: pointer;
}

.sheet-select:disabled {
  background-color: #f5f5f5;
  color: #6b7280;
  cursor: not-allowed;
  border-color: #ccc;
}

.sheet-select:focus {
  outline: none;
  border-color: #34a853;
  box-shadow: 0 0 0 2px rgba(52, 168, 83, 0.1);
}

.sheet-select:disabled:focus {
  border-color: #ccc;
  box-shadow: none;
}
</style>
