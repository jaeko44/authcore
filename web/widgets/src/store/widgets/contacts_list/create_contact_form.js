import cloneDeep from 'lodash-es/cloneDeep'

import {
  CREATE_CONTACT_SPINNER_STATE,
  CREATE_CONTACT_STARTED,
  CREATE_CONTACT_COMPLETED,
  CREATE_CONTACT_FAILED,
  CLEAR_STATES
} from '@/store/types'
import { marshalContact } from '@/utils/marshal'

function getDefaultState () {
  return {
    createdContact: undefined,
    loading: false,
    error: undefined,
    done: false
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    [CREATE_CONTACT_SPINNER_STATE] (state, loading) {
      state.spinnerLoading = loading
    },
    [CREATE_CONTACT_STARTED] (state) {
      state.loading = true
      state.done = false
    },
    [CREATE_CONTACT_COMPLETED] (state, contact) {
      state.createdContact = cloneDeep(marshalContact(contact))
      state.loading = false
      state.done = true
      state.error = undefined
    },
    [CREATE_CONTACT_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
      if (err.response !== undefined) {
        switch (err.response.status) {
          case 400:
            state.error = 'contact_create.input.error.invalid_contact'
            break
          case 409:
            state.error = 'contact_create.input.error.duplicate_contact'
            break
          case 429:
            state.error = 'contact_create.input.error.too_frequent'
            break
        }
      }
    },

    // Clear states
    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async create ({ commit, dispatch, rootState: { client = undefined } }, { type, value }) {
      try {
        commit(CREATE_CONTACT_STARTED)
        let contact
        switch (type) {
          case 'EMAIL':
            contact = await client.authcoreClient.updateCurrentUser({
              primary_email: value
            })
            await client.authcoreClient.startVerifyPrimaryContact('email')
            break
          case 'PHONE':
            contact = await client.authcoreClient.updateCurrentUser({
              primary_phone: value
            })
            await client.authcoreClient.startVerifyPrimaryContact('phone')
            break
          default: throw new Error(`contact type ${type} is not defined`)
        }
        commit(CREATE_CONTACT_COMPLETED, contact)
      } catch (err) {
        commit(CREATE_CONTACT_FAILED, err)
      }
    }
  }
}
