const TOKEN_KEY = 'bajet_auth_token'
const USER_KEY  = 'bajet_auth_user'

let _token = localStorage.getItem(TOKEN_KEY) || ''
let _user  = null
try { _user = JSON.parse(localStorage.getItem(USER_KEY) || 'null') } catch { /**/ }

export function getToken()   { return _token }
export function getUser()    { return _user }
export function isSignedIn() { return !!_token }

export function signIn(credential) {
  _token = credential
  // Decode the JWT payload client-side (no security concern — backend re-verifies every request)
  const [, payloadB64] = credential.split('.')
  const payload = JSON.parse(atob(payloadB64.replace(/-/g, '+').replace(/_/g, '/')))
  _user = { name: payload.name, email: payload.email, picture: payload.picture }
  localStorage.setItem(TOKEN_KEY, _token)
  localStorage.setItem(USER_KEY, JSON.stringify(_user))
}

export function signOut() {
  _token = ''
  _user  = null
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}
