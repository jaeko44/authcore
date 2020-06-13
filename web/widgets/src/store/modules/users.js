import client from '@/client'

export default {
  namespaced: true,

  state: {
    currentUser: {},
    displayName: ''
  },

  actions: {
    async getCurrentUser ({ commit }) {
      try {
        const currentUser = await client.client.getCurrentUser()
        commit('SET_CURRENT_USER', currentUser)
      } catch (e) {
        console.error(e)
      }
    }
  },

  mutations: {
    SET_CURRENT_USER (state, user) {
      state.currentUser = user
      if (user) {
        state.displayName = user.name || user.preferred_username || user.email || user.phone_number || ''
      } else {
        state.displayName = ''
      }
    }
  }
}
