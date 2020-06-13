import cloneDeep from 'lodash/cloneDeep'

import client from '@/client'
import { marshalUser } from '../../../utils/marshal'

function getDefaultState () {
  return {
    loading: false,
    listDone: false,
    error: undefined,
    updateDone: false,
    updateError: undefined,

    userMetadata: '',
    appMetadata: ''
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    SET_METADATA (state, payload) {
      const { userMetadata, appMetadata } = payload
      state.userMetadata = cloneDeep(userMetadata)
      state.appMetadata = cloneDeep(appMetadata)
    },
    SET_LOADING (state) {
      state.loading = true
      state.error = undefined
      state.updateDone = false
      state.updateError = undefined
    },
    UNSET_LOADING (state) {
      state.loading = false
    },
    SET_LIST_DONE (state) {
      state.listDone = true
    },
    SET_UPDATE_DONE (state) {
      state.updateDone = true
    },
    SET_ERROR (state, err) {
      console.log(err)
      state.loading = false
      state.error = err
    },

    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async list ({ commit }, user) {
      try {
        commit('SET_LOADING')
        commit('SET_METADATA', {
          userMetadata: user.userMetadata !== undefined ? user.userMetadata : '',
          appMetadata: user.appMetadata !== undefined ? user.appMetadata : ''
        })
        commit('UNSET_LOADING')
        commit('SET_LIST_DONE')
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async updateUserMetadata ({ commit }, payload) {
      const { id, userMetadata } = payload
      try {
        commit('SET_LOADING')
        JSON.parse(userMetadata) // will throw when `userMetadata` is not a valid JSON object
        const res = await client.client.updateUser(id, { user_metadata: userMetadata })
        const user = marshalUser(res)
        commit('SET_METADATA', {
          userMetadata: user.userMetadata !== undefined ? user.userMetadata : '',
          appMetadata: user.appMetadata !== undefined ? user.appMetadata : ''
        })
        commit('UNSET_LOADING')
        commit('SET_UPDATE_DONE')
        commit('management/alert/SET_MESSAGE', {
          type: 'success',
          message: 'user_details_metadata.message.update_metadata',
          pending: false
        }, { root: true })
      } catch (err) {
        commit('management/alert/SET_MESSAGE', {
          type: 'danger',
          message: 'error.invalid_user_metadata_syntax',
          pending: false
        }, { root: true })
      }
    },
    async updateAppMetadata ({ commit }, payload) {
      const { id, appMetadata } = payload
      try {
        commit('SET_LOADING')
        JSON.parse(appMetadata) // will throw when `appMetadata` is not a valid JSON object
        const res = await client.client.updateUser(id, { app_metadata: appMetadata })
        const user = marshalUser(res)
        commit('SET_METADATA', {
          userMetadata: user.userMetadata !== undefined ? user.userMetadata : '',
          appMetadata: user.appMetadata !== undefined ? user.appMetadata : ''
        })
        commit('UNSET_LOADING')
        commit('SET_UPDATE_DONE')
        commit('management/alert/SET_MESSAGE', {
          type: 'success',
          message: 'user_details_metadata.message.update_metadata',
          pending: false
        }, { root: true })
      } catch (err) {
        commit('management/alert/SET_MESSAGE', {
          type: 'danger',
          message: 'error.invalid_app_metadata_syntax',
          pending: false
        }, { root: true })
      }
    }
  }
}
