<template>
  <div class="task-card" :class="{ running: task.status === 'running' }">
    <!-- Edit mode -->
    <template v-if="editing">
      <div class="task-edit-form">
        <input v-model="editTitle" class="edit-input" placeholder="Title" @keydown.enter="saveEdit" @keydown.escape="cancelEdit" />
        <textarea v-model="editDesc" class="edit-textarea" placeholder="Description (optional)" rows="2" @keydown.escape="cancelEdit" />
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
          <div class="task-title">{{ task.title }}</div>
          <div v-if="task.description" class="task-description">{{ task.description }}</div>
        </div>
        <div class="task-actions">
          <template v-if="task.status === 'pending'">
            <button class="btn-primary" @click="$emit('start', task.id)">Start</button>
          </template>
          <template v-else-if="task.status === 'running'">
            <button class="btn-ghost" @click="$emit('pause', task.id)">Pause</button>
            <button class="btn-ghost" @click="$emit('finish', task.id)">Finish</button>
            <button class="btn-ghost" @click="$emit('open-timer', task.id)">Timer</button>
          </template>
          <template v-else-if="task.status === 'paused'">
            <button class="btn-primary" @click="$emit('start', task.id)">Resume</button>
            <button class="btn-ghost" @click="$emit('finish', task.id)">Finish</button>
          </template>
          <template v-else-if="task.status === 'finished'">
            <button class="btn-ghost" @click="$emit('start', task.id)">Continue</button>
          </template>
          <button class="btn-icon" title="Edit" @click="startEdit">✎</button>
          <button class="btn-danger" @click="$emit('delete', task.id)">✕</button>
        </div>
      </div>
      <div class="task-meta">
        <span class="elapsed">{{ formatElapsed(task.elapsedMs) }}</span>
        <span class="status-badge" :class="task.status">{{ task.status }}</span>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  task: { type: Object, required: true },
})

const emit = defineEmits(['start', 'pause', 'finish', 'edit', 'delete', 'open-timer'])

const editing = ref(false)
const editTitle = ref('')
const editDesc = ref('')

function startEdit() {
  editTitle.value = props.task.title
  editDesc.value = props.task.description ?? ''
  editing.value = true
}

function saveEdit() {
  const title = editTitle.value.trim()
  if (!title) return
  emit('edit', { id: props.task.id, title, description: editDesc.value.trim() })
  editing.value = false
}

function cancelEdit() {
  editing.value = false
}

function formatElapsed(ms) {
  if (!ms || ms < 0) ms = 0
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}
</script>
