// Supabase client — used when VITE_SUPABASE_URL and VITE_SUPABASE_ANON_KEY are
// set at build time. The browser talks to Supabase directly; row level
// security and the functions in supabase/schema.sql protect the data.
import { createClient } from '@supabase/supabase-js'

const URL = import.meta.env.VITE_SUPABASE_URL
const KEY = import.meta.env.VITE_SUPABASE_ANON_KEY

export const supabase = URL && KEY
  ? createClient(URL, KEY, { auth: { persistSession: true, autoRefreshToken: true } })
  : null

// supabase-js restores the session asynchronously; this synchronous peek at its
// storage lets the first render skip the sign-in screen for returning users.
export function hasStoredSession() {
  if (!supabase) return false
  try {
    const ref = new globalThis.URL(URL).hostname.split('.')[0]
    return !!localStorage.getItem(`sb-${ref}-auth-token`)
  } catch { return false }
}

async function sha256Hex(text) {
  const buf = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text))
  return [...new Uint8Array(buf)].map(b => b.toString(16).padStart(2, '0')).join('')
}

// Renders the Google button into `el`. Google gets the hashed nonce and
// Supabase the raw one, so a captured ID token can't be replayed elsewhere.
export async function renderGoogleButton(el, onSignedIn, onError) {
  const nonce = crypto.randomUUID()
  const hashed = await sha256Hex(nonce)
  window.google.accounts.id.initialize({
    client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
    nonce: hashed,
    callback: async ({ credential }) => {
      const { error } = await supabase.auth.signInWithIdToken({ provider: 'google', token: credential, nonce })
      if (error) onError?.(error)
      else onSignedIn()
    },
  })
  window.google.accounts.id.renderButton(el, { theme: 'outline', size: 'large', shape: 'pill' })
}

export function supabaseUser(session) {
  const u = session?.user
  if (!u) return null
  const m = u.user_metadata ?? {}
  return { name: m.full_name ?? m.name ?? u.email, email: u.email, picture: m.avatar_url ?? m.picture }
}
