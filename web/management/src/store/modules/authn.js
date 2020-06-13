import client from '@/client'

export default {
  namespaced: true,
  state () {
    return {
      loading: false,
      error: undefined
    }
  },

  getters: {
    isAuthenticated () {
      return !!client.tokenManager.get('access_token')
    }
  },

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    SET_ERROR (state, err) {
      state.error = err
    }
  },

  actions: {
    async createAccessToken ({ commit }, authorizationToken) {
      try {
        commit('SET_LOADING')
        const webUrl = new URL(document.location)
        webUrl.searchParams.delete('code')
        webUrl.searchParams.delete('state')
        const token = await client.client.exchange('authorization_code', authorizationToken, {
          redirect_uri: webUrl.toString()
        })
        client.tokenManager.add('access_token', token.access_token)
      } catch (err) {
        console.error(`error creating access token ${err}`)
        commit('SET_ERROR', err)
        commit('management/alert/SET_MESSAGE', {
          type: 'danger',
          message: 'error.unknown',
          pending: false
        }, { root: true })
      }
    }
  }
}
