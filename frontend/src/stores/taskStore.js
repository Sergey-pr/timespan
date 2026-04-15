import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { Events } from '@wailsio/runtime'
import {
  GetTasks,
  CreateTask,
  StartTask,
  PauseTask,
  FinishTask,
  DeleteTask,
  EditTask,
  OpenTimerWindow,
  CloseTimerWindow,
} from '../../bindings/timespan/app.js'

export const useTaskStore = defineStore('tasks', () => {
  const tasks = ref([])

  const activeTasks = computed(() =>
    tasks.value.filter(t => t.status !== 'finished')
  )

  const finishedTasks = computed(() =>
    tasks.value.filter(t => t.status === 'finished')
  )

  const runningTask = computed(() =>
    tasks.value.find(t => t.status === 'running') ?? null
  )

  async function loadTasks() {
    tasks.value = await GetTasks()
    tasks.value.forEach(resetSegment)
  }

  async function createTask(title, description) {
    const task = await CreateTask(title, description)
    if (task) {
      tasks.value.unshift(task)
      resetSegment(task)
    }
    return task
  }

  async function startTask(id) {
    const updated = await StartTask(id)
    if (updated) applyUpdate(updated)
    return updated
  }

  async function pauseTask(id) {
    const updated = await PauseTask(id)
    if (updated) applyUpdate(updated)
    return updated
  }

  async function finishTask(id) {
    const updated = await FinishTask(id)
    if (updated) applyUpdate(updated)
    return updated
  }

  async function editTask(id, title, description) {
    const updated = await EditTask(id, title, description)
    if (updated) applyUpdate(updated)
    return updated
  }

  async function deleteTask(id) {
    const ok = await DeleteTask(id)
    if (ok) tasks.value = tasks.value.filter(t => t.id !== id)
    return ok
  }

  function openTimer(id) {
    OpenTimerWindow(id)
  }

  function closeTimer() {
    CloseTimerWindow()
  }

  // Advance elapsedMs in memory for running tasks on every tick.
  function tick() {
    const now = Date.now()
    tasks.value.forEach(t => {
      if (t.status === 'running' && t._segStart != null) {
        t.elapsedMs = t._baseElapsed + (now - t._segStart)
      }
    })
  }

  function resetSegment(task) {
    task._baseElapsed = task.elapsedMs
    task._segStart = task.startedAt ? new Date(task.startedAt).getTime() : null
  }

  function applyUpdate(updated) {
    resetSegment(updated)
    const idx = tasks.value.findIndex(t => t.id === updated.id)
    if (idx !== -1) {
      tasks.value[idx] = updated
    } else {
      // task moved back from finished to active (e.g. restarted)
      tasks.value.unshift(updated)
    }
  }

  function setupTicker() {
    Events.On('tick', tick)
    Events.On('task:updated', (ev) => { if (ev.data) applyUpdate(ev.data) })
  }

  return {
    tasks,
    activeTasks,
    finishedTasks,
    runningTask,
    loadTasks,
    createTask,
    startTask,
    pauseTask,
    finishTask,
    editTask,
    deleteTask,
    openTimer,
    closeTimer,
    setupTicker,
  }
})
