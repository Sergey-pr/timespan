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
  GetCategories,
  CreateCategory,
  RenameCategory,
  DeleteCategory,
  ExportReport,
} from '../../bindings/timespan/app.js'
import { TaskStatus } from '../constants/taskStatus.js'
import { segmentStart, liveElapsed } from '../utils/time.js'

export const SortMode = Object.freeze({
  RecentActivity: 'recent_activity',
  CreatedDesc:    'created_desc',
  CreatedAsc:     'created_asc',
  TitleAsc:       'title_asc',
  ElapsedDesc:    'elapsed_desc',
  Status:         'status',
})

export const SortModeLabel = Object.freeze({
  [SortMode.RecentActivity]: 'Recent activity',
  [SortMode.CreatedDesc]:    'Newest first',
  [SortMode.CreatedAsc]:     'Oldest first',
  [SortMode.TitleAsc]:       'By title A–Z',
  [SortMode.ElapsedDesc]:    'Most time first',
  [SortMode.Status]:         'By status',
})

const STATUS_PRIORITY = {
  [TaskStatus.StatusActive]:       0,
  [TaskStatus.StatusPaused]:       1,
  [TaskStatus.StatusReadyToStart]: 2,
  [TaskStatus.StatusFinished]:     3,
}

function taskComparator(mode) {
  switch (mode) {
    case SortMode.RecentActivity:
      return (a, b) => {
        const ta = a.startedAt ? new Date(a.startedAt) : new Date(0)
        const tb = b.startedAt ? new Date(b.startedAt) : new Date(0)
        return tb - ta
      }
    case SortMode.CreatedAsc:
      return (a, b) => new Date(a.createdAt) - new Date(b.createdAt)
    case SortMode.TitleAsc:
      return (a, b) => (a.title ?? '').localeCompare(b.title ?? '')
    case SortMode.ElapsedDesc:
      return (a, b) => (b.elapsedMs ?? 0) - (a.elapsedMs ?? 0)
    case SortMode.Status:
      return (a, b) => (STATUS_PRIORITY[a.status] ?? 9) - (STATUS_PRIORITY[b.status] ?? 9)
    default:
      return (a, b) => new Date(b.createdAt) - new Date(a.createdAt)
  }
}

export const CategorySortMode = Object.freeze({
  NewestFirst:  'newest_first',
  OldestFirst:  'oldest_first',
  NameAsc:      'name_asc',
  MostTasks:    'most_tasks',
})

export const CategorySortModeLabel = Object.freeze({
  [CategorySortMode.NewestFirst]: 'Newest first',
  [CategorySortMode.OldestFirst]: 'Oldest first',
  [CategorySortMode.NameAsc]:     'A–Z',
  [CategorySortMode.MostTasks]:   'Most tasks',
})

const SORT_STORAGE_KEY = 'timespan:sortMode'
const CAT_SORT_STORAGE_KEY = 'timespan:categorySortMode'

function loadSortMode() {
  try {
    const v = localStorage.getItem(SORT_STORAGE_KEY)
    return Object.values(SortMode).includes(v) ? v : SortMode.RecentActivity
  } catch { return SortMode.RecentActivity }
}

function loadCategorySortMode() {
  try {
    const v = localStorage.getItem(CAT_SORT_STORAGE_KEY)
    return Object.values(CategorySortMode).includes(v) ? v : CategorySortMode.NewestFirst
  } catch { return CategorySortMode.NewestFirst }
}

export const useTaskStore = defineStore('tasks', () => {
  const tasks = ref([])
  const categories = ref([])
  const searchTerm = ref('')
  const sortMode = ref(loadSortMode())
  const categorySortMode = ref(loadCategorySortMode())

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
    tasks.value.filter(t => t.status !== TaskStatus.StatusFinished && matchesSearch(t))
  )

  const finishedTasks = computed(() =>
    tasks.value.filter(t => t.status === TaskStatus.StatusFinished && matchesSearch(t))
  )

  const runningTask = computed(() =>
    tasks.value.find(t => t.status === TaskStatus.StatusActive) ?? null
  )

  // Positions are decided when the list itself changes, so starting or ticking
  // a task never moves it.
  const order = ref(new Map())

  function refreshOrder() {
    const sorted = [...tasks.value].sort(taskComparator(sortMode.value))
    order.value = new Map(sorted.map((t, i) => [t.id, i]))
  }

  function orderOf(task) {
    return order.value.get(task.id) ?? Number.MAX_SAFE_INTEGER
  }

  function setSortMode(mode) {
    sortMode.value = mode
    try { localStorage.setItem(SORT_STORAGE_KEY, mode) } catch {}
    refreshOrder()
  }

  function setCategorySortMode(mode) {
    categorySortMode.value = mode
    try { localStorage.setItem(CAT_SORT_STORAGE_KEY, mode) } catch {}
  }

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
    for (const group of map.values()) {
      group.tasks.sort((a, b) => orderOf(a) - orderOf(b))
    }
    return [...map.values()].sort((a, b) => {
      if (!a.category && !b.category) return 0
      if (!a.category) return 1
      if (!b.category) return -1
      switch (categorySortMode.value) {
        case CategorySortMode.OldestFirst:
          return new Date(a.category.createdAt) - new Date(b.category.createdAt)
        case CategorySortMode.NameAsc:
          return a.category.name.localeCompare(b.category.name)
        case CategorySortMode.MostTasks:
          return b.tasks.length - a.tasks.length
        default:
          return new Date(b.category.createdAt) - new Date(a.category.createdAt)
      }
    })
  }

  const activeByCategory = computed(() => groupByCategory(activeTasks.value))
  const finishedByCategory = computed(() => groupByCategory(finishedTasks.value))

  async function loadTasks() {
    tasks.value = await GetTasks()
    tasks.value.forEach(resetSegment)
    refreshOrder()
  }

  async function loadCategories() {
    categories.value = await GetCategories()
  }

  async function createCategory(name) {
    const cat = await CreateCategory(name)
    if (cat) {
      categories.value.unshift(cat)
    }
    return cat
  }

  async function renameCategory(id, name) {
    const cat = await RenameCategory(id, name)
    if (cat) {
      const idx = categories.value.findIndex(c => c.id === id)
      if (idx !== -1) categories.value[idx] = cat
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
      refreshOrder()
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
    if (updated) {
      applyUpdate(updated)
      refreshOrder()
    }
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

  // Advance elapsedMs in memory for running tasks on every tick.
  function tick() {
    const now = Date.now()
    tasks.value.forEach(t => {
      if (t.status === TaskStatus.StatusActive && t._segStart != null) {
        t.elapsedMs = liveElapsed(t._baseElapsed, t._segStart, now)
      }
    })
  }

  function resetSegment(task) {
    task._baseElapsed = task.elapsedMs
    task._segStart = segmentStart(task)
  }

  function applyUpdate(updated) {
    resetSegment(updated)
    const idx = tasks.value.findIndex(t => t.id === updated.id)
    if (idx !== -1) {
      tasks.value[idx] = updated
    } else {
      // task moved back from finished to active (e.g. restarted)
      tasks.value.unshift(updated)
      refreshOrder()
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
    sortMode,
    setSortMode,
    categorySortMode,
    setCategorySortMode,
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
    setupTicker,
  }
})
