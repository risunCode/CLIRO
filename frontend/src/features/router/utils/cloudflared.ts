export type CloudflaredMode = 'quick' | 'auth'

export interface CloudflaredModeCard {
  id: CloudflaredMode
  label: string
  description: string
}

export const CLOUDFLARED_MODE_CARDS: CloudflaredModeCard[] = [
  {
    id: 'quick',
    label: 'Quick Tunnel',
    description: 'Auto-generated temporary URL (*.trycloudflare.com), no account needed.'
  },
  {
    id: 'auth',
    label: 'Named Tunnel',
    description: 'Use a Cloudflare tunnel token for persistent public access or custom domains.'
  }
]
