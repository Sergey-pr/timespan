<template>
  <Teleport to="body">
    <div class="modal-backdrop" @click.self="$emit('close')">
      <div class="modal categories-modal" @keydown.escape="$emit('close')">

        <div class="modal-header">
          <span class="modal-title">Categories</span>
          <button class="btn-icon modal-close" @click="$emit('close')">✕</button>
        </div>

        <div class="modal-body">

          <!-- Existing categories -->
          <div v-if="store.categories.length" class="cat-list">
            <div
              v-for="cat in store.categories"
              :key="cat.id"
              class="cat-row"
            >
              <!-- Rename mode -->
              <template v-if="renamingId === cat.id">
                <input
                  :ref="el => { if (el) renameInputRef = el }"
                  v-model="renameValue"
                  class="cat-rename-input"
                  @keydown.enter.prevent="saveRename"
                  @keydown.escape="cancelRename"
                />
                <button class="btn-icon cat-action" title="Save" @click="saveRename">✓</button>
                <button class="btn-icon cat-action" title="Cancel" @click="cancelRename">✕</button>
              </template>

              <!-- View mode -->
              <template v-else>
                <span class="cat-name">{{ cat.name }}</span>
                <div class="cat-actions">
                  <button class="btn-icon cat-action" title="Rename" @click="startRename(cat)">✎</button>
                  <button
                    class="btn-icon cat-action cat-delete"
                    :class="{ disabled: isUsed(cat.id) }"
                    :title="isUsed(cat.id) ? 'Category has tasks' : 'Delete'"
                    :disabled="isUsed(cat.id)"
                    @click="handleDelete(cat.id)"
                  >✕</button>
                </div>
              </template>
            </div>
          </div>

          <div v-else class="cat-empty">No categories yet</div>

          <!-- New category input -->
          <div class="cat-new-row">
            <input
              ref="newCategoryInputRef"
              v-model="newName"
              class="cat-new-input"
              placeholder="New category name…"
              maxlength="64"
              @keydown.enter.prevent="handleCreate"
            />
            <button
              class="btn-primary cat-add-btn"
              :disabled="!newName.trim()"
              @click="handleCreate"
            >Add</button>
          </div>

        </div>

        <div class="modal-footer">
          <button class="btn-primary" @click="$emit('close')">Done</button>
        </div>

      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import { useTaskStore } from '../stores/taskStore.js'

const emit = defineEmits(['close'])

const store = useTaskStore()

// Set of category IDs that have at least one task
const usedIds = computed(() =>
  new Set(store.tasks.map(t => t.categoryId).filter(Boolean))
)
function isUsed(id) { return usedIds.value.has(id) }

// New category
const newName = ref('')
const newCategoryInputRef = ref(null)

async function handleCreate() {
  const name = newName.value.trim()
  if (!name) return
  await store.createCategory(name)
  newName.value = ''
  newCategoryInputRef.value?.focus()
}

// Rename
const renamingId = ref(null)
const renameValue = ref('')
let renameInputRef = null

async function startRename(cat) {
  renamingId.value = cat.id
  renameValue.value = cat.name
  await nextTick()
  renameInputRef?.focus()
  renameInputRef?.select()
}

async function saveRename() {
  const name = renameValue.value.trim()
  if (name && name !== store.categories.find(c => c.id === renamingId.value)?.name) {
    await store.renameCategory(renamingId.value, name)
  }
  renamingId.value = null
}

function cancelRename() {
  renamingId.value = null
}

// Delete
async function handleDelete(id) {
  if (isUsed(id)) return
  await store.deleteCategory(id)
}
</script>
