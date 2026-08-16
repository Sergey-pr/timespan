<template>
  <div class="timer-window">
    <div class="timer-header">
      <span class="timer-title" :title="task?.title">{{ task?.title ?? '-' }}</span>
      <button class="timer-close" @click="close">✕</button>
    </div>
    <div class="timer-elapsed">{{ formattedElapsed }}</div>
    <div class="timer-actions">
      <button v-if="task?.status === TaskStatus.ACTIVE" @click="pause">Pause</button>
      <button v-else-if="task?.status === TaskStatus.PAUSED" @click="resume">Resume</button>
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
import { TaskStatus } from './constants/taskStatus.js'
import { formatElapsed, segmentStart, liveElapsed } from './utils/time.js'

const task = ref(null)
const baseElapsed = ref(0)
const segmentStartedAt = ref(null)
const now = ref(Date.now())

const taskId = parseInt(new URLSearchParams(window.location.search).get('taskId') ?? '0', 10)

const formattedElapsed = computed(() => {
  const running = task.value?.status === TaskStatus.ACTIVE
  return formatElapsed(
    liveElapsed(baseElapsed.value, running ? segmentStartedAt.value : null, now.value),
  )
})

async function loadTask() {
  if (!taskId) return
  const tasks = await GetTasks()
  const found = tasks.find(t => t.id === taskId)
  if (found) {
    task.value = found
    baseElapsed.value = found.elapsedMs
    segmentStartedAt.value = segmentStart(found)
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
  segmentStartedAt.value = segmentStart(updated)
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
    segmentStartedAt.value = segmentStart(updated) ?? Date.now()
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
