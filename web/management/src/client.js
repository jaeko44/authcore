import { Authcore } from 'authcore-js'

export default new Authcore({
  clientId: '_authcore_admin_portal_',
  baseURL: window.origin
})
