import type { AsyncTaskState } from '@/shared/lib/async'
import type { SettingsViewState } from '@/features/settings/types'

export const deriveSettingsViewState = (task: AsyncTaskState, statusMessage: string): SettingsViewState => {
  return {
    busy: task.busy,
    statusMessage,
    errorMessage: task.error
  }
}
