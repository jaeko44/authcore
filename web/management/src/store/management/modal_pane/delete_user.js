import client from '@/client'

export default {
  namespaced: true,
  modules: {},

  state: {
    loading: false,
    done: false,
    error: undefined
  },
  getters: {},

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    SET_DONE (state) {
      state.loading = false
      state.done = true
    },
    SET_ERROR (state, error) {
      console.error(state)
      state.loading = false
      state.error = error
    },

    RESET_STATES (state) {
      state.loading = false
      state.done = false
      state.error = undefined
    }
  },

  actions: {
    async delete ({ commit, rootState: { currentUser } }, userId) {
      try {
        commit('SET_LOADING')
        const { id: currentUserId } = currentUser.user
        if (currentUserId === userId) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'alert.error.delete_user_for_same_account',
            pending: true
          }, { root: true })
          throw new Error('alert.error.delete_user_for_same_account')
        }
        await client.client.deleteUser(userId)
        commit('management/alert/SET_MESSAGE', {
          type: 'success',
          message: 'user_list.message.user_deleted_successfully',
          pending: true
        }, { root: true })
        commit('SET_DONE')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
