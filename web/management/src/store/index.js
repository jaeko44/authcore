import Vue from 'vue'
import Vuex from 'vuex'

import authn from './modules/authn'
import currentUser from './current_user'
import management from './management'

Vue.use(Vuex)

export default new Vuex.Store({
  modules: {
    authn,
    currentUser,
    management
  }
})
