const TOKEN_KEY = 'bajet_auth_token'
const USER_KEY  = 'bajet_auth_user'

let _token = localStorage.getItem(TOKEN_KEY) || ''
let _user  = null
try { _user = JSON.parse(localStorage.getItem(USER_KEY) || 'null') } catch { /**/ }

export function getToken()   { return _token }
export function getUser()    { return _user }
export function isSignedIn() { return !!_token }

// userInfo is optional — if provided (backend mode) we use it directly;
// otherwise we decode the JWT payload client-side (localStorage fallback mode).
export function signIn(token, userInfo) {
  _token = token
  if (userInfo) {
    _user = { name: userInfo.name, email: userInfo.email, picture: userInfo.picture }
  } else {
    try {
      const [, p] = token.split('.')
      const payload = JSON.parse(atob(p.replace(/-/g, '+').replace(/_/g, '/')))
      _user = { name: payload.name, email: payload.email, picture: payload.picture }
    } catch { _user = null }
  }
  localStorage.setItem(TOKEN_KEY, _token)
  localStorage.setItem(USER_KEY, JSON.stringify(_user))
}

export function signOut() {
  _token = ''
  _user  = null
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}
