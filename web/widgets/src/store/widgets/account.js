import cloneDeep from 'lodash-es/cloneDeep'

import {
  SET_PROFILE_COMPLETED,
  GET_PROFILE_STARTED,
  GET_PROFILE_COMPLETED,
  GET_PROFILE_FAILED,

  CLEAR_STATES
} from '@/store/types'

import { loadLanguageAsync } from '@/i18n-setup'
import { marshalUser } from '@/utils/marshal'

import verifyAccountForm from './account/verify_account_form'
import resetPasswordForm from './account/reset_password_form'
import updateProfileForm from './account/update_profile_form'

function getDefaultState () {
  return {
    user: {
      profileName: '',
      displayName: '',
      username: '',
      primaryEmail: '',
      primaryEmailVerified: undefined,
      primaryPhone: '',
      primaryPhoneVerified: undefined,
      smsAuthentication: undefined,
      totpAuthentication: undefined,
      passwordAuthentication: undefined
    },
    loading: false,
    done: false,
    error: undefined
  }
}

export default {
  namespaced: true,
  modules: {
    verifyAccountForm,
    resetPasswordForm,
    updateProfileForm
  },

  state: getDefaultState(),
  getters: {
    pageLoaded: (state) => {
      return state.done || state.error !== undefined
    }
  },

  mutations: {
    [SET_PROFILE_COMPLETED] (state, user) {
      state.user = cloneDeep(marshalUser(user))
      if (user.language !== '') {
        loadLanguageAsync(user.language)
      }
    },

    [GET_PROFILE_STARTED] (state) {
      state.loading = true
      state.done = false
    },
    [GET_PROFILE_COMPLETED] (state) {
      state.loading = false
      state.done = true
      state.error = undefined
    },
    [GET_PROFILE_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    // Clear states
    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async get ({ state, commit, rootState: { client = undefined } }) {
      try {
        commit(GET_PROFILE_STARTED)
        const currentUser = await client.authcoreClient.getCurrentUser()
        commit(GET_PROFILE_COMPLETED)
        commit(SET_PROFILE_COMPLETED, currentUser)
      } catch (err) {
        commit(GET_PROFILE_FAILED, err)
      }
    }
  }
}
