// Formats milliseconds as HH:MM:SS.
export function formatElapsed(ms) {
  if (!ms || ms < 0) ms = 0
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

// Timestamp the task's running segment started at, or null when it is not running.
// Go clears started_at on pause and finish, so null means there is nothing to advance.
export function segmentStart(task) {
  return task?.startedAt ? new Date(task.startedAt).getTime() : null
}

// Elapsed time including the segment currently in progress, if any.
export function liveElapsed(baseMs, segmentStartedAt, now = Date.now()) {
  if (segmentStartedAt == null) return baseMs
  return baseMs + (now - segmentStartedAt)
}
