import { randomTOTPSecret } from 'authcore-js/src/crypto/random'

const REDIRECT_PATH_KEY = 'io.authcore.redirect_path'
const OAUTH_STATE_KEY = 'io.authcore.admin_portal_oauth_state'

// setRedirectPath function set the redirect path to be stored in session storage
export function setRedirectPath (redirectPath) {
  sessionStorage.setItem(REDIRECT_PATH_KEY, redirectPath)
}

// getRedirectPath function return the redirect path stored in session storage
export function getRedirectPath () {
  return sessionStorage.getItem(REDIRECT_PATH_KEY)
}

// removeRedirectPath function remove the redirect path stored in session storage
export function removeRedirectPath () {
  sessionStorage.removeItem(REDIRECT_PATH_KEY)
}

// Create OAuth state
export function createOAuthState () {
  const state = randomTOTPSecret().toString('utf-8')
  sessionStorage.setItem(OAUTH_STATE_KEY, state)
  return state
}

// Verify OAuth state
export function verifyOAuthState (state) {
  const stored = sessionStorage.getItem(OAUTH_STATE_KEY)
  if (state === stored) {
    sessionStorage.removeItem(OAUTH_STATE_KEY)
    return true
  }
  return false
}
