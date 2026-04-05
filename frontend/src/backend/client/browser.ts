import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'

export const openBrowserURL = (url: string): void => {
  BrowserOpenURL(url)
}
