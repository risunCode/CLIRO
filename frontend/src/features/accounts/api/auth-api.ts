import type { auth } from '../../../../wailsjs/go/models'
import { CancelCodexAuth, CancelKiroAuth, GetCodexAuthSession, GetKiroAuthSession, StartCodexAuth, StartKiroAuth, StartKiroSocialAuth } from '@/shared/api/wails/client'
import type { AuthSession, KiroAuthSession } from '@/features/accounts/types'

export const accountsAuthApi = {
  startCodexAuth: (): Promise<auth.CodexAuthStart> => StartCodexAuth(),
  getCodexAuthSession: (sessionId: string): Promise<AuthSession> => GetCodexAuthSession(sessionId),
  cancelCodexAuth: (sessionId: string): Promise<void> => CancelCodexAuth(sessionId),
  startKiroAuth: (): Promise<auth.KiroAuthStart> => StartKiroAuth(),
  startKiroSocialAuth: (provider: string): Promise<auth.KiroAuthStart> => StartKiroSocialAuth(provider),
  getKiroAuthSession: (sessionId: string): Promise<KiroAuthSession> => GetKiroAuthSession(sessionId),
  cancelKiroAuth: (sessionId: string): Promise<void> => CancelKiroAuth(sessionId)
}
