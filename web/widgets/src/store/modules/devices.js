import client from '@/client'

function getDefaultState () {
  return {
    sessions: [],
    currentSession: undefined,

    loading: false,
    done: false,
    error: null
  }
}

export default {
  namespaced: true,

  state: getDefaultState(),
  getters: {
    pageLoaded: (state) => {
      return state.done || state.error !== undefined
    }
  },

  mutations: {
    SET_LOADING (state) {
      state.loading = true
      state.error = null
    },
    UNSET_LOADING (state) {
      state.loading = false
    },
    SET_DEVICES_LIST (state, { sessions, currentSession }) {
      state.sessions = sessions.sort(function (a, b) {
        if (a.id === currentSession.id) return -1
        if (b.id === currentSession.id) return 1
        if (a.last_seen_at > b.last_seen_at) {
          return -1
        }
        if (a.last_seen_at < b.last_seen_at) {
          return 1
        }
        return 0
      })
      state.currentSession = currentSession
      state.loading = false
      state.done = true
    },
    SET_ERROR (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    RESET (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async list ({ commit }) {
      try {
        commit('SET_LOADING')
        const resp = await client.client.listCurrentUserSessions()
        const currentSession = await client.client.getCurrentSession()
        commit('SET_DEVICES_LIST', { sessions: resp.results, currentSession })
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async delete ({ commit }, deviceId) {
      try {
        commit('SET_LOADING')
        await client.client.deleteCurrentUserSession(deviceId)
        commit('UNSET_LOADING')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async deleteAll ({ state, commit }) {
      try {
        commit('SET_LOADING')
        await Promise.all(state.sessions.map(async (device) => {
          if (device.id !== state.currentSession.id) {
            await client.client.deleteCurrentUserSession(device.id)
          }
        }))
        commit('UNSET_LOADING')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
