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
            @start="store.startTask($event)"
            @pause="store.pauseTask($event)"
            @finish="store.finishTask($event)"
            @edit="e => store.editTask(e.id, e.title, e.description)"
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
            @start="store.startTask($event)"
            @edit="e => store.editTask(e.id, e.title, e.description)"
            @delete="store.deleteTask($event)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useTaskStore } from './stores/taskStore.js'
import TaskCard from './components/TaskCard.vue'

const store = useTaskStore()
const newTitle = ref('')
const newDescription = ref('')
const doneOpen = ref(false)

async function handleCreate() {
  const title = newTitle.value.trim()
  if (!title) return
  await store.createTask(title, newDescription.value.trim())
  newTitle.value = ''
  newDescription.value = ''
}

onMounted(async () => {
  await store.loadTasks()
  store.setupTicker()
})
</script>
