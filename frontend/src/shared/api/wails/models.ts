import type { codex, config, kiro, logger, main } from '../../../../wailsjs/go/models'

export type WailsAppState = main.State
export type WailsLogEntry = logger.Entry
export type WailsAccount = config.Account

export type WailsCodexAuthStart = codex.AuthStart
export type WailsCodexAuthSessionView = codex.AuthSessionView
export type WailsKiroAuthStart = kiro.AuthStart
