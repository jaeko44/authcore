import { Authcore } from 'authcore-js'

const urlParams = new URLSearchParams(window.location.search)
const clientId = urlParams.get('clientId')

export default new Authcore({
  clientId: clientId,
  baseURL: window.origin
})
