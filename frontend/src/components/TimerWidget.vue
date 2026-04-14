<template>
  <div v-if="visible" class="timer-overlay">
    <div class="timer-overlay-header">
      <span class="timer-overlay-title" :title="task?.title">{{ task?.title ?? '—' }}</span>
      <button class="timer-overlay-close" @click="close">✕</button>
    </div>
    <div class="timer-overlay-elapsed">{{ formattedElapsed }}</div>
    <div class="timer-overlay-actions">
      <button
        v-if="task?.status === 'running'"
        class="btn-ghost"
        @click="togglePause"
      >Pause</button>
      <button
        v-else-if="task?.status === 'paused'"
        class="btn-primary"
        @click="togglePause"
      >Resume</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useTaskStore } from '../stores/taskStore.js'

const store = useTaskStore()
const visible = ref(false)
const taskId = ref(null)

const task = computed(() =>
  taskId.value ? store.tasks.find(t => t.id === taskId.value) ?? null : null
)

const formattedElapsed = computed(() => {
  const ms = task.value?.elapsedMs ?? 0
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

function open(id) {
  taskId.value = id
  visible.value = true
}

function close() {
  visible.value = false
  taskId.value = null
  store.closeTimer()
}

async function togglePause() {
  if (!taskId.value) return
  if (task.value?.status === 'running') {
    await store.pauseTask(taskId.value)
  } else {
    await store.startTask(taskId.value)
  }
}

defineExpose({ open, close })
</script>
