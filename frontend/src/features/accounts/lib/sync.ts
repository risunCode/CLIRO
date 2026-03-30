import type { AccountSyncResult, CodexAuthSyncResult, KiloAuthSyncResult, OpencodeAuthSyncResult, SyncTargetID } from '@/features/accounts/types'

interface AccountSyncTarget {
  id: SyncTargetID
  name: string
  path: string
  description: string
}

interface AccountSyncHandlers {
  toKilo: (accountId: string) => Promise<KiloAuthSyncResult>
  toOpencode: (accountId: string) => Promise<OpencodeAuthSyncResult>
  toCodex: (accountId: string) => Promise<CodexAuthSyncResult>
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
  if (target === 'codex-cli') {
    return handlers.toCodex(accountId)
  }

  if (target === 'opencode-cli') {
    return handlers.toOpencode(accountId)
  }

  return handlers.toKilo(accountId)
}
