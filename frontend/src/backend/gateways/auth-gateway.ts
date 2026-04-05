import { wailsClient } from '@/backend/client/wails-client'
import type { WailsAuthSessionView, WailsAuthStart } from '@/backend/models/wails'

export const accountsAuthApi = {
	startAuth: (provider: string): Promise<WailsAuthStart> => wailsClient.auth.startAuth(provider),
	startSocialAuth: (provider: string, socialProvider: string): Promise<WailsAuthStart> => wailsClient.auth.startSocialAuth(provider, socialProvider),
	getAuthSession: (provider: string, sessionId: string): Promise<WailsAuthSessionView> => wailsClient.auth.getAuthSession(provider, sessionId),
	cancelAuth: (provider: string, sessionId: string): Promise<void> => wailsClient.auth.cancelAuth(provider, sessionId),
	submitAuthCode: (provider: string, sessionId: string, code: string): Promise<void> => wailsClient.auth.submitAuthCode(provider, sessionId, code)
}
