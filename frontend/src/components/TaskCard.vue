<template>
  <div class="task-card" :class="{ running: task.status === TaskStatus.ACTIVE }">
    <!-- Edit mode -->
    <template v-if="editing">
      <div class="task-edit-form">
        <input v-model="editTitle" class="edit-input" placeholder="Title" @keydown.enter="saveEdit" @keydown.escape="cancelEdit" />
        <textarea v-model="editDesc" class="edit-textarea" placeholder="Description (optional)" rows="2" @keydown.escape="cancelEdit" />

        <!-- Category row in edit mode -->
        <div class="category-row">
          <select v-model="editCategoryId" class="category-select">
            <option :value="null">No category</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
        </div>

        <div class="edit-actions">
          <button class="btn-primary" @click="saveEdit">Save</button>
          <button class="btn-ghost" @click="cancelEdit">Cancel</button>
        </div>
      </div>
    </template>

    <!-- View mode -->
    <template v-else>
      <div class="task-card-header">
        <div class="task-info">
          <div class="task-title-row">
            <span class="task-title">{{ task.title }}</span>
            <button class="btn-copy" :class="{ copied }" title="Copy title" @click="copyTitle">
              {{ copied ? '✓' : '⎘' }}
            </button>
          </div>
          <div v-if="task.description" class="task-description-row">
            <span class="task-description">{{ task.description }}</span>
            <button class="btn-copy" :class="{ copied: copiedDesc }" title="Copy description" @click="copyDescription">
              {{ copiedDesc ? '✓' : '⎘' }}
            </button>
          </div>
        </div>
        <div class="task-actions">
          <template v-if="task.status === TaskStatus.READY_TO_START">
            <button class="btn-primary" @click="$emit('start', task.id)">Start</button>
          </template>
          <template v-else-if="task.status === TaskStatus.ACTIVE">
            <button class="btn-ghost" @click="$emit('pause', task.id)">Pause</button>
            <button class="btn-ghost" @click="$emit('finish', task.id)">Finish</button>
            <button class="btn-ghost" @click="$emit('open-timer', task.id)">Timer</button>
          </template>
          <template v-else-if="task.status === TaskStatus.PAUSED">
            <button class="btn-primary" @click="$emit('start', task.id)">Resume</button>
            <button class="btn-ghost" @click="$emit('finish', task.id)">Finish</button>
          </template>
          <template v-else-if="task.status === TaskStatus.FINISHED">
            <button class="btn-ghost" @click="$emit('start', task.id)">Continue</button>
          </template>
          <button class="btn-icon" title="Edit" @click="startEdit">✎</button>
          <template v-if="confirmingDelete">
            <span class="confirm-delete-label">Delete?</span>
            <button class="btn-danger-solid" @click="$emit('delete', task.id)">Yes</button>
            <button class="btn-ghost" @click="confirmingDelete = false">No</button>
          </template>
          <button v-else class="btn-danger" @click="confirmingDelete = true">✕</button>
        </div>
      </div>
      <div class="task-meta">
        <span class="elapsed">{{ formatElapsed(task.elapsedMs) }}</span>
        <span class="status-badge" :class="task.status">{{ statusLabel }}</span>
        <span v-if="categoryName" class="category-badge">{{ categoryName }}</span>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { TaskStatus, TaskStatusLabel } from '../constants/taskStatus.js'
import { formatElapsed } from '../utils/time.js'

const props = defineProps({
  task:       { type: Object, required: true },
  categories: { type: Array,  default: () => [] },
})

const emit = defineEmits(['start', 'pause', 'finish', 'edit', 'delete', 'open-timer'])

const editing = ref(false)
const confirmingDelete = ref(false)
const copied = ref(false)
const copiedDesc = ref(false)
const editTitle = ref('')
const editDesc = ref('')
const editCategoryId = ref(null)

const statusLabel = computed(() => TaskStatusLabel[props.task.status] ?? props.task.status)

const categoryName = computed(() => {
  if (!props.task.categoryId) return null
  return props.categories.find(c => c.id === props.task.categoryId)?.name ?? null
})

function startEdit() {
  editTitle.value = props.task.title
  editDesc.value = props.task.description ?? ''
  editCategoryId.value = props.task.categoryId ?? null
  editing.value = true
}

function saveEdit() {
  const title = editTitle.value.trim()
  if (!title) return
  emit('edit', {
    id:          props.task.id,
    title,
    description: editDesc.value.trim(),
    categoryId:  editCategoryId.value,
  })
  editing.value = false
}

function cancelEdit() {
  editing.value = false
}

function copyTitle() {
  navigator.clipboard.writeText(props.task.title)
  copied.value = true
  setTimeout(() => { copied.value = false }, 1200)
}

function copyDescription() {
  navigator.clipboard.writeText(props.task.description)
  copiedDesc.value = true
  setTimeout(() => { copiedDesc.value = false }, 1200)
}
</script>
