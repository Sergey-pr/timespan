import { TaskStatus } from '../../bindings/timespan/models.js'

// Values come from the Go constants via the generated bindings and cannot drift.
export { TaskStatus }

// Human-readable labels (mirrors statusLabels in export.go).
export const TaskStatusLabel = Object.freeze({
  [TaskStatus.StatusReadyToStart]: 'Ready to start',
  [TaskStatus.StatusActive]:       'Active',
  [TaskStatus.StatusPaused]:       'Paused',
  [TaskStatus.StatusFinished]:     'Finished',
})
