// Canonical task status values — must match the TaskStatus constants in task.go
// byte-for-byte. This is the single source of truth on the frontend; compare
// task.status against TaskStatus.* rather than string literals.
export const TaskStatus = Object.freeze({
  READY_TO_START: 'ready_to_start',
  ACTIVE:         'active',
  PAUSED:         'paused',
  FINISHED:       'finished',
})

// Human-readable labels for display (mirrors statusLabels in export.go).
export const TaskStatusLabel = Object.freeze({
  [TaskStatus.READY_TO_START]: 'Ready to start',
  [TaskStatus.ACTIVE]:         'Active',
  [TaskStatus.PAUSED]:         'Paused',
  [TaskStatus.FINISHED]:       'Finished',
})
