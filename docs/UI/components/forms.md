# Forms

## Pattern Terbaik

**Best Example:** [UserCreateForm.vue](../../../frontend/src/components/UserCreateForm.vue) & [UserEditForm.vue](../../../frontend/src/components/UserEditForm.vue)

### Structure Pattern

```vue
<form @submit.prevent="submitForm" class="form">
  <div class="form-group">
    <label for="field-id">Field Name</label>
    <input id="field-id" v-model="form.field" type="text" required />
  </div>
  
  <div v-if="message" :class="['message', status]">
    {{ message }}
  </div>
  
  <div class="form-actions">
    <button type="submit" class="btn btn-primary">Submit</button>
    <button type="button" class="btn btn-secondary">Cancel</button>
  </div>
</form>
```

---

## Base Styling

```css
.form-container {
  max-width: 700px; /* Settings forms */
  /* OR max-width: 500px for modals */
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
  font-size: 14px;
  color: #374151;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  font-family: inherit;
  transition: all 0.2s;
}
```

---

## Focus States

```css
.form-group input:focus,
.form-group select:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}
```

---

## Disabled States

```css
.form-group input:disabled,
.form-group select:disabled {
  background: #f5f5f5;
  cursor: not-allowed;
  opacity: 0.6;
}
```

---

## Helper Text

```css
.form-group small {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #6b7280;
}
```

```vue
<div class="form-group">
  <label for="email">Email Address</label>
  <input id="email" type="email" v-model="form.email" />
  <small>We'll never share your email with anyone else.</small>
</div>
```

---

## Validation Messages

```css
.message {
  margin-bottom: 15px;
  padding: 12px;
  border-radius: 4px;
  font-weight: 500;
}

.message.success {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.message.error {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.message.warning {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffeeba;
}
```

---

## Form Actions

```css
.form-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
  justify-content: flex-end;
}
```

---

## Input Types

### Text Input

```vue
<div class="form-group">
  <label for="username">Username</label>
  <input id="username" type="text" v-model="form.username" required />
</div>
```

### Select Dropdown

```vue
<div class="form-group">
  <label for="role">Role</label>
  <select id="role" v-model="form.role" required>
    <option value="">-- Select Role --</option>
    <option value="admin">Admin</option>
    <option value="user">User</option>
  </select>
</div>
```

### Textarea

```vue
<div class="form-group">
  <label for="description">Description</label>
  <textarea id="description" v-model="form.description" rows="4"></textarea>
</div>
```

```css
textarea {
  resize: vertical;
  min-height: 80px;
}
```

### Checkbox

```vue
<div class="form-group">
  <label class="checkbox-label">
    <input type="checkbox" v-model="form.agreed" />
    <span>I agree to the terms and conditions</span>
  </label>
</div>
```

```css
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: auto;
  cursor: pointer;
}
```

---

## Accessibility Features

- ✅ `<label for="id">` properly connected
- ✅ `required` attribute for validation
- ✅ `aria-label` for icon buttons
- ✅ Disabled state styling
- ✅ Error messages associated with inputs

---

## Responsive

```css
@media (max-width: 768px) {
  .form-container {
    padding: 16px;
  }

  .form-actions {
    flex-direction: column;
  }

  .form-actions button {
    width: 100%;
  }
}
```

---

## Quick Reference

| Property          | Value                               |
| ----------------- | ----------------------------------- |
| **Max Width**     | `700px` (settings)                  |
| **Input Padding** | `10px 12px`                         |
| **Input Border**  | `1px solid #d1d5db`                 |
| **Border Radius** | `6px`                               |
| **Focus Border**  | `#3b82f6`                           |
| **Focus Shadow**  | `0 0 0 3px rgba(59, 130, 246, 0.1)` |
| **Gap**           | `16px` (form-group)                 |

---

**See Also:** [Buttons](./buttons.md) | [Accessibility](../accessibility.md)
