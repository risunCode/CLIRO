import { EventsOff, EventsOn } from '../../../wailsjs/runtime/runtime'

export type RuntimeEventUnsubscribe = () => void

export const subscribeRuntimeEvent = (
  eventName: string,
  onEvent: (payload: unknown) => void
): RuntimeEventUnsubscribe => {
  return EventsOn(eventName, (payload: unknown) => {
    onEvent(payload)
  })
}

export const unsubscribeRuntimeEvent = (eventName: string): void => {
  EventsOff(eventName)
}
