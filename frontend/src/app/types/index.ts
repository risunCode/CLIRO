// Re-export app-level types from backend/models so app/ code can still import from here
export type { AppState, LogEntry, UpdateInfo } from '@/backend/models/system'
