# Modals

## Pattern Terbaik

**Best Example:** [Modal.vue](../../../frontend/src/components/Modal.vue) (Generic reusable modal)

---

## Structure

```vue
<teleport to="body">
  <transition name="modal-fade">
    <div v-if="isOpen" class="modal-backdrop" @click="closeOnBackdrop">
      <div class="modal-container" @click.stop>
        <div class="modal-header">
          <h2 class="modal-title">{{ title }}</h2>
          <button class="modal-close-btn" aria-label="Close modal" @click="close">
            <i class="pi pi-times"></i>
          </button>
        </div>
        <div class="modal-body">
          <slot />
        </div>
        <div class="modal-footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </transition>
</teleport>
```

---

## Styling

### Backdrop

```css
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
}
```

### Container

```css
.modal-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 600px; /* Default, bisa override */
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #e5e7eb;
}
```

### Header

```css
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
  background: linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%);
}

.modal-title {
  font-size: 18px;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
}

.modal-close-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  color: #6b7280;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 200ms ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-close-btn:hover {
  background-color: #e5e7eb;
  color: #3b82f6;
}
```

### Body

```css
.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  color: #1f2937;
}

/* Custom scrollbar */
.modal-body::-webkit-scrollbar {
  width: 6px;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.modal-body::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}
```

### Footer

```css
.modal-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  padding: 16px 24px;
  border-top: 1px solid #e5e7eb;
  background: linear-gradient(135deg, #f9fafb 0%, #f3f4f6 100%);
}
```

---

## Transitions

```css
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: all 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from .modal-container,
.modal-fade-leave-to .modal-container {
  transform: scale(0.95);
}
```

---

## Body Scroll Lock

```javascript
// In component script
watch(
  () => props.isOpen,
  (isOpen) => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
  },
);

// Cleanup on unmount
onUnmounted(() => {
  document.body.style.overflow = "";
});
```

---

## Usage Examples

### Basic Modal

```vue
<Modal :is-open="isOpen" title="Confirm Action" @close="isOpen = false">
  <p>Are you sure you want to delete this item?</p>
  
  <template #footer>
    <button class="btn btn-secondary" @click="isOpen = false">Cancel</button>
    <button class="btn btn-danger" @click="confirmDelete">Delete</button>
  </template>
</Modal>
```

### Form Modal

```vue
<Modal :is-open="showForm" title="Create User" @close="showForm = false">
  <form @submit.prevent="submitForm">
    <div class="form-group">
      <label for="username">Username</label>
      <input id="username" v-model="form.username" type="text" required />
    </div>
    
    <div class="form-group">
      <label for="email">Email</label>
      <input id="email" v-model="form.email" type="email" required />
    </div>
  </form>
  
  <template #footer>
    <button class="btn btn-secondary" @click="showForm = false">Cancel</button>
    <button class="btn btn-primary" @click="submitForm">Create</button>
  </template>
</Modal>
```

### Large Modal (Override Width)

```vue
<Modal
  :is-open="isOpen"
  title="Details"
  @close="isOpen = false"
  class="modal-large"
>
  <!-- Content -->
</Modal>
```

```css
.modal-large .modal-container {
  max-width: 900px;
}
```

---

## Modal Sizes

```css
/* Small Modal */
.modal-sm .modal-container {
  max-width: 400px;
}

/* Medium Modal (Default) */
.modal-container {
  max-width: 600px;
}

/* Large Modal */
.modal-lg .modal-container {
  max-width: 900px;
}

/* Extra Large Modal */
.modal-xl .modal-container {
  max-width: 1200px;
}
```

---

## Responsive

```css
@media (max-width: 640px) {
  .modal-container {
    max-width: 95%;
    max-height: 95vh;
    border-radius: 8px;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 16px;
  }

  .modal-title {
    font-size: 16px;
  }
}
```

---

## Accessibility

- ✅ `teleport to="body"` to avoid z-index issues
- ✅ `aria-label` on close button
- ✅ Click backdrop to close (optional)
- ✅ ESC key to close
- ✅ Focus trap within modal
- ✅ Body scroll lock when open
- ✅ Restore scroll on close

---

## ESC Key Handler

```javascript
onMounted(() => {
  const handleEsc = (e) => {
    if (e.key === "Escape" && props.isOpen) {
      emit("close");
    }
  };
  window.addEventListener("keydown", handleEsc);
  onUnmounted(() => {
    window.removeEventListener("keydown", handleEsc);
  });
});
```

---

## Quick Reference

| Property          | Value                         |
| ----------------- | ----------------------------- |
| **Max Width**     | `600px` (default)             |
| **Max Height**    | `90vh`                        |
| **Border Radius** | `12px`                        |
| **Shadow**        | `0 20px 60px rgba(0,0,0,0.3)` |
| **Backdrop**      | `rgba(0,0,0,0.5)` + blur      |
| **Z-index**       | `9999`                        |
| **Padding**       | `24px` (body)                 |

---

**See Also:** [Buttons](./buttons.md) | [Forms](./forms.md) | [Accessibility](../accessibility.md)
