import { asBoolean, asNumber, asRecord, asString, asStringArray, pick } from '@/backend/compat/coerce'
import type { CodexAuthSyncResult, KiloAuthSyncResult, OpencodeAuthSyncResult, SyncResultBase } from '@/features/accounts/types'

const toSyncResultBase = (payload: unknown): SyncResultBase => {
  const record = asRecord(payload)

  return {
    targetPath: asString(pick(record, 'targetPath', 'target_path')),
    fileExisted: asBoolean(pick(record, 'fileExisted', 'file_existed')),
    updatedFields: asStringArray(pick(record, 'updatedFields', 'updated_fields')),
    accountID: asString(pick(record, 'accountID', 'account_id')),
    provider: asString(record.provider)
  }
}

const toOAuthSyncResult = (
  payload: unknown,
  target: 'kilo-cli' | 'opencode-cli'
): KiloAuthSyncResult | OpencodeAuthSyncResult => {
  const record = asRecord(payload)
  const base = toSyncResultBase(payload)
  const result = {
    ...base,
    target,
    openAICreated: asBoolean(pick(record, 'openAICreated', 'openai_created')),
    syncedExpires: asNumber(pick(record, 'syncedExpires', 'synced_expires')),
    syncedExpiresAt: asString(pick(record, 'syncedExpiresAt', 'synced_expires_at')) || undefined
  }

  if (target === 'kilo-cli') {
    return result as KiloAuthSyncResult
  }

  return result as OpencodeAuthSyncResult
}

export const toKiloAuthSyncResult = (payload: unknown): KiloAuthSyncResult => {
  return toOAuthSyncResult(payload, 'kilo-cli') as KiloAuthSyncResult
}

export const toOpencodeAuthSyncResult = (payload: unknown): OpencodeAuthSyncResult => {
  return toOAuthSyncResult(payload, 'opencode-cli') as OpencodeAuthSyncResult
}

export const toCodexAuthSyncResult = (payload: unknown): CodexAuthSyncResult => {
  const record = asRecord(payload)

  return {
    ...toSyncResultBase(record),
    target: 'codex-cli',
    backupPath: asString(pick(record, 'backupPath', 'backup_path')) || undefined,
    backupCreated: asBoolean(pick(record, 'backupCreated', 'backup_created')),
    syncedAt: asString(pick(record, 'syncedAt', 'synced_at'))
  }
}
