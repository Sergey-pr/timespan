<template>
  <div id="app-root">
    <div class="titlebar">
      <h1>TimeSpan</h1>
    </div>

    <div class="main-content">
      <!-- Add task form -->
      <form class="add-task-form" @submit.prevent="handleCreate">
        <input
          v-model="newTitle"
          placeholder="New task title…"
          autocomplete="off"
          required
        />
        <textarea
          v-model="newDescription"
          placeholder="Description (optional)"
        />

        <!-- Category row -->
        <div class="category-row">
          <select v-model="newCategoryId" class="category-select">
            <option :value="null">No category</option>
            <option v-for="cat in store.categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
        </div>
        <div class="form-actions">
          <button type="button" class="btn-primary" @click="showCategoryModal = true">Edit categories</button>
          <button type="submit" class="btn-primary">Add task</button>
        </div>
      </form>

      <!-- Active tasks grouped by category -->
      <div v-if="store.activeTasks.length">
        <div class="section-header">Tasks</div>
        <template v-for="group in store.activeByCategory" :key="group.category?.id ?? 0">
          <!-- Collapsible header only when grouping is visible -->
          <button
            v-if="store.activeByCategory.length > 1 || group.category"
            class="category-group-toggle"
            @click="toggleGroup(groupKey(group, 'active'))"
          >
            <span class="chevron" :class="{ open: isGroupOpen(group, 'active') }">›</span>
            {{ group.category ? group.category.name : 'No category' }}
            <span class="group-count">({{ group.tasks.length }})</span>
          </button>
          <div v-if="isGroupOpen(group, 'active')" class="task-list">
            <TaskCard
              v-for="task in group.tasks"
              :key="task.id"
              :task="task"
              :categories="store.categories"
              @start="store.startTask($event)"
              @pause="store.pauseTask($event)"
              @finish="store.finishTask($event)"
              @edit="e => store.editTask(e.id, e.title, e.description, e.categoryId)"
              @delete="store.deleteTask($event)"
              @open-timer="store.openTimer($event)"
            />
          </div>
        </template>
      </div>

      <!-- Finished tasks (collapsible) grouped by category -->
      <div v-if="store.finishedTasks.length">
        <button class="done-toggle" @click="doneOpen = !doneOpen">
          <span class="chevron" :class="{ open: doneOpen }">›</span>
          Done ({{ store.finishedTasks.length }})
        </button>
        <template v-if="doneOpen" v-for="group in store.finishedByCategory" :key="group.category?.id ?? 0">
          <button
            v-if="store.finishedByCategory.length > 1 || group.category"
            class="category-group-toggle"
            @click="toggleGroup(groupKey(group, 'finished'))"
          >
            <span class="chevron" :class="{ open: isGroupOpen(group, 'finished') }">›</span>
            {{ group.category ? group.category.name : 'No category' }}
            <span class="group-count">({{ group.tasks.length }})</span>
          </button>
          <div v-if="isGroupOpen(group, 'finished')" class="task-list">
            <TaskCard
              v-for="task in group.tasks"
              :key="task.id"
              :task="task"
              :categories="store.categories"
              @start="store.startTask($event)"
              @edit="e => store.editTask(e.id, e.title, e.description, e.categoryId)"
              @delete="store.deleteTask($event)"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- Categories management modal -->
    <CategoriesModal
      v-if="showCategoryModal"
      @close="showCategoryModal = false"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useTaskStore } from './stores/taskStore.js'
import TaskCard from './components/TaskCard.vue'
import CategoriesModal from './components/CategoriesModal.vue'

const store = useTaskStore()
const newTitle = ref('')
const newDescription = ref('')
const newCategoryId = ref(null)
const doneOpen = ref(false)
const showCategoryModal = ref(false)
const collapsedGroups = ref(new Set())

function toggleGroup(key) {
  if (collapsedGroups.value.has(key)) {
    collapsedGroups.value.delete(key)
  } else {
    collapsedGroups.value.add(key)
  }
  // trigger reactivity
  collapsedGroups.value = new Set(collapsedGroups.value)
}

function groupKey(group, prefix) {
  return `${prefix}:${group.category?.id ?? 0}`
}

function isGroupOpen(group, prefix) {
  return !collapsedGroups.value.has(groupKey(group, prefix))
}

async function handleCreate() {
  const title = newTitle.value.trim()
  if (!title) return
  await store.createTask(title, newDescription.value.trim(), newCategoryId.value)
  newTitle.value = ''
  newDescription.value = ''
  newCategoryId.value = null
}


onMounted(async () => {
  await Promise.all([store.loadTasks(), store.loadCategories()])
  store.setupTicker()
})
</script>
