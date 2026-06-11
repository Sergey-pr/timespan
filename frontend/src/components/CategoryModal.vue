<template>
  <Teleport to="body">
    <div class="modal-backdrop" @click.self="$emit('close')">
      <div class="modal" @keydown.escape="$emit('close')">
        <div class="modal-header">
          <span class="modal-title">New category</span>
          <button class="btn-icon modal-close" @click="$emit('close')">✕</button>
        </div>

        <div class="modal-body">
          <input
            ref="inputRef"
            v-model="name"
            class="modal-input"
            placeholder="Category name"
            maxlength="64"
            @keydown.enter.prevent="submit"
            @keydown.escape="$emit('close')"
          />
        </div>

        <div class="modal-footer">
          <button class="btn-ghost" @click="$emit('close')">Cancel</button>
          <button class="btn-primary" :disabled="!name.trim()" @click="submit">Create</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const emit = defineEmits(['close', 'create'])

const name = ref('')
const inputRef = ref(null)

onMounted(() => inputRef.value?.focus())

function submit() {
  const trimmed = name.value.trim()
  if (!trimmed) return
  emit('create', trimmed)
  emit('close')
}
</script>
