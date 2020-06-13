import cloneDeep from 'lodash-es/cloneDeep'

import {
  LIST_CONTACTS_STARTED,
  LIST_CONTACTS_COMPLETED,
  LIST_CONTACTS_FAILED,

  CLEAR_STATES,
  CLEAR_ACTION_STATES
} from '@/store/types'
import { marshalContact } from '@/utils/marshal'

import createContactForm from './contacts_list/create_contact_form'
import deleteContactForm from './contacts_list/delete_contact_form'
import updatePrimaryContactForm from './contacts_list/update_primary_contact_form'

function getDefaultState () {
  return {
    contacts: [],
    loading: false,
    done: false,
    error: undefined,
    deleteDone: false,
    deleteError: undefined,
    updatePrimaryDone: false,
    updatePrimaryError: undefined
  }
}

export default {
  namespaced: true,
  modules: {
    createContactForm,
    deleteContactForm,
    updatePrimaryContactForm
  },

  state: getDefaultState(),
  getters: {},

  mutations: {
    // List contacts
    [LIST_CONTACTS_STARTED] (state) {
      state.loading = true
      state.done = false
    },
    [LIST_CONTACTS_COMPLETED] (state, payload) {
      const { contacts, isMarshalNeeded } = payload
      if (isMarshalNeeded) {
        state.contacts = cloneDeep(contacts.map(marshalContact).sort(function (a, b) {
          // Sort contact to let primary be the first one
          return !!b.isPrimary - !!a.isPrimary
        }))
      } else {
        state.contacts = cloneDeep(contacts)
      }
      state.loading = false
      state.done = true
      state.error = undefined
    },
    [LIST_CONTACTS_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    // Clear action states
    [CLEAR_ACTION_STATES] (state) {
      state.resendDone = false
      state.resendError = undefined
      state.deleteDone = false
      state.deleteError = undefined
      state.verifyDone = false
      state.verifyError = undefined
      state.updatePrimaryDone = false
      state.updatePrimaryError = undefined
    },

    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async list ({ commit, rootState: { client = undefined } }, payload) {
      const { type = '' } = payload
      try {
        commit(LIST_CONTACTS_STARTED)
        const { contacts } = await client.authcoreClient.listContacts(type)
        commit(CLEAR_ACTION_STATES)
        commit(LIST_CONTACTS_COMPLETED, {
          contacts: contacts || [],
          isMarshalNeeded: true
        })
      } catch (err) {
        commit(CLEAR_ACTION_STATES)
        commit(LIST_CONTACTS_FAILED, err)
      }
    }
  }
}
