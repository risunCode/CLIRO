import type { AccountSyncResult, SyncTargetID } from '@/features/accounts/types'

interface AccountSyncTarget {
  id: SyncTargetID
  name: string
  path: string
  description: string
}

interface AccountSyncHandlers {
  sync: (accountId: string, target: SyncTargetID) => Promise<AccountSyncResult>
}

export const ACCOUNT_SYNC_TARGETS: readonly AccountSyncTarget[] = [
  {
    id: 'kilo-cli',
    name: 'Kilo CLI',
    path: '~/.local/share/kilo/auth.json',
    description: 'Sync this Codex account into the Kilo CLI auth file.'
  },
  {
    id: 'opencode-cli',
    name: 'Opencode',
    path: '~/.local/share/opencode/auth.json',
    description: 'Sync this Codex account into the Opencode auth file.'
  },
  {
    id: 'codex-cli',
    name: 'Codex CLI',
    path: '~/.codex/auth.json',
    description: 'Sync this Codex account into the Codex CLI auth file.'
  }
]

export const syncTargetName = (target: SyncTargetID): string => {
  return ACCOUNT_SYNC_TARGETS.find((item) => item.id === target)?.name || 'Kilo CLI'
}

export const runAccountSyncByTarget = async (
  accountId: string,
  target: SyncTargetID,
  handlers: AccountSyncHandlers
): Promise<AccountSyncResult> => {
  return handlers.sync(accountId, target)
}
