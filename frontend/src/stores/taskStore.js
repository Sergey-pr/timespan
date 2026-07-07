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
  GetCategories,
  CreateCategory,
  RenameCategory,
  DeleteCategory,
  ExportReport,
} from '../../bindings/timespan/app.js'
import { TaskStatus } from '../constants/taskStatus.js'

export const useTaskStore = defineStore('tasks', () => {
  const tasks = ref([])
  const categories = ref([])
  const searchTerm = ref('')

  // Case-insensitive substring match on title or description.
  // Done in memory (not via SQL LIKE/ILIKE) because SQLite's LIKE is ASCII-only
  // and the store already holds every task for the live timers.
  function matchesSearch(t) {
    const term = searchTerm.value.trim().toLowerCase()
    if (!term) return true
    const title = (t.title ?? '').toLowerCase()
    const description = (t.description ?? '').toLowerCase()
    return title.includes(term) || description.includes(term)
  }

  const activeTasks = computed(() =>
    tasks.value.filter(t => t.status !== TaskStatus.FINISHED && matchesSearch(t))
  )

  const finishedTasks = computed(() =>
    tasks.value.filter(t => t.status === TaskStatus.FINISHED && matchesSearch(t))
  )

  const runningTask = computed(() =>
    tasks.value.find(t => t.status === TaskStatus.ACTIVE) ?? null
  )

  // Groups a task list by category: named categories sorted A–Z, uncategorised last.
  // Returns [{ category: Category|null, tasks: Task[] }, ...]
  function groupByCategory(taskList) {
    const map = new Map()
    for (const task of taskList) {
      const key = task.categoryId ?? 0
      if (!map.has(key)) {
        const cat = key ? (categories.value.find(c => c.id === key) ?? null) : null
        map.set(key, { category: cat, tasks: [] })
      }
      map.get(key).tasks.push(task)
    }
    return [...map.values()].sort((a, b) => {
      if (!a.category && !b.category) return 0
      if (!a.category) return 1
      if (!b.category) return -1
      return a.category.name.localeCompare(b.category.name)
    })
  }

  const activeByCategory = computed(() => groupByCategory(activeTasks.value))
  const finishedByCategory = computed(() => groupByCategory(finishedTasks.value))

  async function loadTasks() {
    tasks.value = await GetTasks()
    tasks.value.forEach(resetSegment)
  }

  async function loadCategories() {
    categories.value = await GetCategories()
  }

  async function createCategory(name) {
    const cat = await CreateCategory(name)
    if (cat) {
      categories.value.push(cat)
      categories.value.sort((a, b) => a.name.localeCompare(b.name))
    }
    return cat
  }

  async function renameCategory(id, name) {
    const cat = await RenameCategory(id, name)
    if (cat) {
      const idx = categories.value.findIndex(c => c.id === id)
      if (idx !== -1) categories.value[idx] = cat
      categories.value.sort((a, b) => a.name.localeCompare(b.name))
    }
    return cat
  }

  async function deleteCategory(id) {
    const ok = await DeleteCategory(id)
    if (ok) categories.value = categories.value.filter(c => c.id !== id)
    return ok
  }

  async function createTask(title, description, categoryId) {
    const task = await CreateTask(title, description, categoryId ?? 0)
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

  async function editTask(id, title, description, categoryId) {
    const updated = await EditTask(id, title, description, categoryId ?? 0)
    if (updated) applyUpdate(updated)
    return updated
  }

  async function deleteTask(id) {
    const ok = await DeleteTask(id)
    if (ok) tasks.value = tasks.value.filter(t => t.id !== id)
    return ok
  }

  async function exportReport() {
    return await ExportReport()
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
      if (t.status === TaskStatus.ACTIVE && t._segStart != null) {
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
    categories,
    searchTerm,
    activeTasks,
    finishedTasks,
    activeByCategory,
    finishedByCategory,
    runningTask,
    loadTasks,
    loadCategories,
    createCategory,
    renameCategory,
    deleteCategory,
    createTask,
    startTask,
    pauseTask,
    finishTask,
    editTask,
    deleteTask,
    exportReport,
    openTimer,
    closeTimer,
    setupTicker,
  }
})
