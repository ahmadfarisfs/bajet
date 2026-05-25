import { writable, derived } from 'svelte/store'

export const sessionExpired = writable(false)

const _count = writable(0)
export const syncing = derived(_count, n => n > 0)

export function startSync() { _count.update(n => n + 1) }
export function endSync()   { _count.update(n => Math.max(0, n - 1)) }
