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
const segStart = ref(null)
const now = ref(Date.now())

const liveElapsed = computed(() => {
  if (task.value?.status === 'running' && segStart.value !== null) {
    return baseElapsed.value + (now.value - segStart.value)
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

async function loadTask(id) {
  const tasks = await GetTasks()
  const found = tasks.find(t => t.id === id)
  if (found) {
    task.value = found
    baseElapsed.value = found.elapsedMs
    segStart.value = found.startedAt ? new Date(found.startedAt).getTime() : null
  }
}

function onTick() {
  now.value = Date.now()
}

function onTimerOpen(ev) {
  const id = ev.data
  if (id) loadTask(id)
}

function onTaskUpdated(ev) {
  const updated = ev.data
  if (!updated || updated.id !== task.value?.id) return
  task.value = updated
  baseElapsed.value = updated.elapsedMs
  segStart.value = updated.startedAt ? new Date(updated.startedAt).getTime() : null
  now.value = Date.now()
}

async function pause() {
  if (!task.value) return
  const updated = await PauseTask(task.value.id)
  if (updated) {
    task.value = updated
    baseElapsed.value = updated.elapsedMs
    segStart.value = null
  }
}

async function resume() {
  if (!task.value) return
  const updated = await StartTask(task.value.id)
  if (updated) {
    task.value = updated
    baseElapsed.value = updated.elapsedMs
    segStart.value = updated.startedAt ? new Date(updated.startedAt).getTime() : Date.now()
  }
}

function close() {
  CloseTimerWindow()
}

let offTick, offTimerOpen, offTaskUpdated
onMounted(() => {
  offTick = Events.On('tick', onTick)
  offTimerOpen = Events.On('timer:open', onTimerOpen)
  offTaskUpdated = Events.On('task:updated', onTaskUpdated)
})

onUnmounted(() => {
  offTick?.()
  offTimerOpen?.()
  offTaskUpdated?.()
})
</script>
