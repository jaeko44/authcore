import Vue from 'vue'
import Vuex from 'vuex'

import alert from './alert'

import client from './client'
import widgets from './widgets'
import authn from './modules/authn'
import mfa from './modules/mfa'
import devices from './modules/devices'
import password from './modules/password'
import socialLogin from './modules/social_login'
import preferences from './modules/preferences'
import users from './modules/users'

Vue.use(Vuex)

export default new Vuex.Store({
  modules: {
    alert,
    client,
    widgets,
    preferences,
    authn,
    mfa,
    devices,
    password,
    socialLogin,
    users
  }
})
