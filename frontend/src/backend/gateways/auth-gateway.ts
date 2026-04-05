import { wailsGateway } from '@/backend/gateways/wails-gateway'
import type { WailsCodexAuthStart, WailsKiroAuthStart } from '@/backend/models/wails'
import type { AuthSession, KiroAuthSession } from '@/features/accounts/types'

export const accountsAuthApi = {
  startCodexAuth: (): Promise<WailsCodexAuthStart> => wailsGateway.auth.startCodexAuth(),
  getCodexAuthSession: (sessionId: string): Promise<AuthSession> => wailsGateway.auth.getCodexAuthSession(sessionId),
  cancelCodexAuth: (sessionId: string): Promise<void> => wailsGateway.auth.cancelCodexAuth(sessionId),
  submitCodexAuthCode: (sessionId: string, code: string): Promise<void> => wailsGateway.auth.submitCodexAuthCode(sessionId, code),
  startKiroAuth: (): Promise<WailsKiroAuthStart> => wailsGateway.auth.startKiroAuth(),
  startKiroSocialAuth: (provider: string): Promise<WailsKiroAuthStart> => wailsGateway.auth.startKiroSocialAuth(provider),
  getKiroAuthSession: (sessionId: string): Promise<KiroAuthSession> => wailsGateway.auth.getKiroAuthSession(sessionId) as Promise<KiroAuthSession>,
  cancelKiroAuth: (sessionId: string): Promise<void> => wailsGateway.auth.cancelKiroAuth(sessionId),
  submitKiroAuthCode: (sessionId: string, code: string): Promise<void> => wailsGateway.auth.submitKiroAuthCode(sessionId, code)
}
