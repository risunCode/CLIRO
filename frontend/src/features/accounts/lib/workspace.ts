import type { Account } from '@/features/accounts/types'
import { areAllVisibleSelected, filterAccounts, groupAccountsByProvider, type ProviderGroup } from '@/features/accounts/lib/account'
import { deriveQuotaDisplayStatus } from '@/features/accounts/lib/account-quota'

interface SessionLike {
  sessionId?: string
  status?: string
}

interface AccountsViewState {
  accountsByProvider: ProviderGroup[]
  filteredAccounts: Account[]
  visibleAccountIds: string[]
  hasVisibleAccounts: boolean
  allVisibleSelected: boolean
  selectedAccounts: Account[]
  selectedEnabledCount: number
  bulkToggleToEnabled: boolean
  exhaustedDisabledCount: number
}

interface AccountsVisibilityFilters {
  showExhausted: boolean
  showDisabled: boolean
}

const defaultVisibilityFilters: AccountsVisibilityFilters = {
  showExhausted: true,
  showDisabled: true
}

const isExhaustedAccount = (account: Account): boolean => {
  return deriveQuotaDisplayStatus(account.quota) === 'exhausted'
}

const isDisabledAccount = (account: Account): boolean => {
  return !account.enabled
}

export const parseImportedAccounts = (raw: unknown): Account[] => {
  if (Array.isArray(raw)) {
    return raw.filter((item) => item && typeof item === 'object') as Account[]
  }

  if (raw && typeof raw === 'object') {
    const payload = raw as Record<string, unknown>
    if (Array.isArray(payload.accounts)) {
      return payload.accounts.filter((item) => item && typeof item === 'object') as Account[]
    }
    return [raw as Account]
  }

  return []
}

export const isPendingAuthSession = (session: SessionLike | null): boolean => {
  return (session?.status ?? '') === 'pending'
}

export const shouldAttachPendingSession = (
  showPrompt: boolean,
  promptSessionID: string,
  session: SessionLike | null
): boolean => {
  return showPrompt && promptSessionID === '' && isPendingAuthSession(session) && Boolean(session?.sessionId)
}

export const shouldDismissPromptAfterSuccess = (
  showPrompt: boolean,
  promptSessionID: string,
  session: SessionLike | null
): boolean => {
  return showPrompt && promptSessionID !== '' && session?.sessionId === promptSessionID && session?.status === 'success'
}

export const computeAccountsViewState = (
  accounts: Account[],
  selectedIds: string[],
  selectedProvider: string,
  searchQuery: string,
  visibilityFilters: AccountsVisibilityFilters = defaultVisibilityFilters
): AccountsViewState => {
  const accountsByProvider = groupAccountsByProvider(accounts)
  const providerAndSearchFilteredAccounts = filterAccounts(accounts, {
    providerId: selectedProvider,
    query: searchQuery
  })

  const filteredAccounts = providerAndSearchFilteredAccounts.filter((account) => {
    if (!visibilityFilters.showDisabled && isDisabledAccount(account)) {
      return false
    }

    if (!visibilityFilters.showExhausted && isExhaustedAccount(account)) {
      return false
    }

    return true
  })

  const exhaustedDisabledCount = providerAndSearchFilteredAccounts.filter((account) => {
    return isDisabledAccount(account) || isExhaustedAccount(account)
  }).length

  const visibleAccountIds = filteredAccounts.map((account) => account.id)
  const hasVisibleAccounts = visibleAccountIds.length > 0
  const allVisibleSelected = areAllVisibleSelected(selectedIds, visibleAccountIds)

  const selectedSet = new Set(selectedIds)
  const selectedAccounts = accounts.filter((account) => selectedSet.has(account.id))
  const selectedEnabledCount = selectedAccounts.filter((account) => account.enabled).length

  return {
    accountsByProvider,
    filteredAccounts,
    visibleAccountIds,
    hasVisibleAccounts,
    allVisibleSelected,
    selectedAccounts,
    selectedEnabledCount,
    bulkToggleToEnabled: selectedIds.length > 0 && selectedEnabledCount !== selectedIds.length,
    exhaustedDisabledCount
  }
}

export const sanitizeSelectedIDs = (selectedIds: string[], accounts: Account[]): string[] => {
  const validAccountIDs = new Set(accounts.map((account) => account.id))
  return selectedIds.filter((id) => validAccountIDs.has(id))
}
