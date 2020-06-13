import client from '@/client'

export default {
  namespaced: true,

  state: {
    loading: false,
    error: null,
    done: false,

    oauthFactors: []
  },

  mutations: {
    SET_LOADING (state) {
      state.loading = true
      state.error = null
    },
    SET_SOCIAL_LOGINS_LIST (state, oauthFactors) {
      state.loading = false
      state.oauthFactors = oauthFactors
      state.done = true
    },
    SET_ERROR (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    }
  },
  actions: {
    async list ({ commit }) {
      try {
        commit('SET_LOADING')
        const socialLogins = (await client.client.listCurrentUserIDP()).results
        commit('SET_SOCIAL_LOGINS_LIST', socialLogins || [])
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async delete ({ commit, dispatch }, id) {
      try {
        commit('SET_LOADING')
        await client.client.deleteCurrentUserIDP(id)
        await dispatch('list')
      } catch (err) {
        if (err.response !== undefined) {
          switch (err.response.status) {
            case 400:
              commit('SET_ERROR', 'error.only_way_to_login')
          }
        } else {
          commit('SET_ERROR', err)
        }
      }
    }
  }
}
