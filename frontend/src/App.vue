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
            <option :value="null">Без категории</option>
            <option v-for="cat in store.categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
          <button type="button" class="btn-add-cat" title="Новая категория" @click="openNewCat">+</button>
        </div>

        <!-- Inline new-category input -->
        <div v-if="showNewCat" class="new-cat-row">
          <input
            ref="catInputRef"
            v-model="newCatName"
            class="new-cat-input"
            placeholder="Название категории…"
            @keydown.enter.prevent="addCategory"
            @keydown.escape="cancelNewCat"
          />
          <button type="button" class="btn-icon" title="Сохранить" @click="addCategory">✓</button>
          <button type="button" class="btn-icon" title="Отмена" @click="cancelNewCat">✕</button>
        </div>

        <div class="form-actions">
          <button type="submit" class="btn-primary">Add task</button>
        </div>
      </form>

      <!-- Active tasks -->
      <div v-if="store.activeTasks.length">
        <div class="section-header">Tasks</div>
        <div class="task-list" style="margin-top:6px">
          <TaskCard
            v-for="task in store.activeTasks"
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
      </div>

      <!-- Finished tasks (collapsible) -->
      <div v-if="store.finishedTasks.length">
        <button class="done-toggle" @click="doneOpen = !doneOpen">
          <span class="chevron" :class="{ open: doneOpen }">›</span>
          Done ({{ store.finishedTasks.length }})
        </button>
        <div v-if="doneOpen" class="task-list" style="margin-top:6px">
          <TaskCard
            v-for="task in store.finishedTasks"
            :key="task.id"
            :task="task"
            :categories="store.categories"
            @start="store.startTask($event)"
            @edit="e => store.editTask(e.id, e.title, e.description, e.categoryId)"
            @delete="store.deleteTask($event)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, nextTick, onMounted } from 'vue'
import { useTaskStore } from './stores/taskStore.js'
import TaskCard from './components/TaskCard.vue'

const store = useTaskStore()
const newTitle = ref('')
const newDescription = ref('')
const newCategoryId = ref(null)
const doneOpen = ref(false)
const showNewCat = ref(false)
const newCatName = ref('')
const catInputRef = ref(null)

async function handleCreate() {
  const title = newTitle.value.trim()
  if (!title) return
  await store.createTask(title, newDescription.value.trim(), newCategoryId.value)
  newTitle.value = ''
  newDescription.value = ''
  newCategoryId.value = null
}

async function openNewCat() {
  showNewCat.value = true
  await nextTick()
  catInputRef.value?.focus()
}

async function addCategory() {
  const name = newCatName.value.trim()
  if (!name) return
  const cat = await store.createCategory(name)
  if (cat) newCategoryId.value = cat.id
  newCatName.value = ''
  showNewCat.value = false
}

function cancelNewCat() {
  newCatName.value = ''
  showNewCat.value = false
}

onMounted(async () => {
  await Promise.all([store.loadTasks(), store.loadCategories()])
  store.setupTicker()
})
</script>
