<template>
  <div class="timer-window">
    <div class="timer-header">
      <span class="timer-title" :title="task?.title">{{ task?.title ?? '-' }}</span>
      <button class="timer-close" @click="close">✕</button>
    </div>
    <div class="timer-elapsed">{{ formattedElapsed }}</div>
    <div class="timer-actions">
      <button v-if="task?.status === 'running'" @click="pause">Pause</button>
      <button v-else-if="task?.status === 'paused'" @click="resume">Resume</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Events } from '@wailsio/runtime'
import {
  GetTasks,
  PauseTask,
  StartTask,
  CloseTimerWindow,
} from '../bindings/timespan/app.js'

const task = ref(null)
const baseElapsed = ref(0)
const segmentStartedAt = ref(null)
const now = ref(Date.now())

const taskId = parseInt(new URLSearchParams(window.location.search).get('taskId') ?? '0', 10)

const liveElapsed = computed(() => {
  if (task.value?.status === 'running' && segmentStartedAt.value !== null) {
    return baseElapsed.value + (now.value - segmentStartedAt.value)
  }
  return baseElapsed.value
})

const formattedElapsed = computed(() => {
  const ms = liveElapsed.value
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

async function loadTask() {
  if (!taskId) return
  const tasks = await GetTasks()
  const found = tasks.find(t => t.id === taskId)
  if (found) {
    task.value = found
    baseElapsed.value = found.elapsedMs
    segmentStartedAt.value = found.startedAt ? new Date(found.startedAt).getTime() : null
  }
}

function onTick() {
  now.value = Date.now()
}

function onTaskUpdated(ev) {
  const updated = ev.data
  if (!updated || updated.id !== taskId) return
  task.value = updated
  baseElapsed.value = updated.elapsedMs
  segmentStartedAt.value = updated.startedAt ? new Date(updated.startedAt).getTime() : null
  now.value = Date.now()
}

async function pause() {
  if (!task.value) return
  const updated = await PauseTask(task.value.id)
  if (updated) {
    task.value = updated
    baseElapsed.value = updated.elapsedMs
    segmentStartedAt.value = null
  }
}

async function resume() {
  if (!task.value) return
  const updated = await StartTask(task.value.id)
  if (updated) {
    task.value = updated
    baseElapsed.value = updated.elapsedMs
    segmentStartedAt.value = updated.startedAt ? new Date(updated.startedAt).getTime() : Date.now()
  }
}

function close() {
  CloseTimerWindow(taskId)
}

let offTick, offTaskUpdated
onMounted(async () => {
  await loadTask()
  offTick = Events.On('tick', onTick)
  offTaskUpdated = Events.On('task:updated', onTaskUpdated)
})

onUnmounted(() => {
  offTick?.()
  offTaskUpdated?.()
})
</script>
