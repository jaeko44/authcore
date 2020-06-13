import {
  VERIFY_ACCOUNT_CONTACT_SPINNER_STATE,
  VERIFY_ACCOUNT_CONTACT_RESET,
  VERIFY_ACCOUNT_CONTACT_INIT,
  VERIFY_ACCOUNT_CONTACT_STARTED,
  VERIFY_ACCOUNT_CONTACT_COMPLETED,
  VERIFY_ACCOUNT_CONTACT_FAILED,

  CREATE_AUTHORIZATION_TOKEN_COMPLETED,
  CREATE_AUTHORIZATION_TOKEN_FAILED,

  RESEND_VERIFY_ACCOUNT_CONTACT_INIT,
  RESEND_VERIFY_ACCOUNT_CONTACT_COMPLETED,
  CLEAR_STATES
} from '@/store/types'

function getDefaultState () {
  return {
    loading: false,
    done: false,
    error: undefined,
    contact: {
      type: undefined,
      value: undefined
    },
    oldContactFlag: undefined,
    spinnerLoading: false,

    createAuthorizationTokenDone: false,
    authorizationToken: undefined,

    resendDone: false
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    [VERIFY_ACCOUNT_CONTACT_SPINNER_STATE] (state, loading) {
      state.spinnerLoading = loading
    },
    [VERIFY_ACCOUNT_CONTACT_RESET] (state) {
      state.error = undefined
    },
    [VERIFY_ACCOUNT_CONTACT_INIT] (state, payload) {
      const { contact, oldContact } = payload
      if (contact.length === 1) {
        state.contact = contact[0]
        state.oldContactFlag = oldContact
      }
    },
    [VERIFY_ACCOUNT_CONTACT_STARTED] (state) {
      state.loading = true
      state.done = false
      state.error = undefined
    },
    [VERIFY_ACCOUNT_CONTACT_COMPLETED] (state) {
      state.loading = false
      state.done = true
    },
    [VERIFY_ACCOUNT_CONTACT_FAILED] (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
      if (err.response !== undefined) {
        switch (err.response.status) {
          case 401:
            state.error = 'verification.input.error.invalid_verification_code'
            break
          case 429:
            state.error = 'verification.input.error.too_frequent'
            break
        }
      }
    },

    [CREATE_AUTHORIZATION_TOKEN_COMPLETED] (state, token) {
      state.createAuthorizationTokenDone = true
      state.authorizationToken = token
    },
    [CREATE_AUTHORIZATION_TOKEN_FAILED] (state, err) {
      console.error(err)
      state.error = err
    },

    [RESEND_VERIFY_ACCOUNT_CONTACT_INIT] (state) {
      state.resendDone = false
    },
    [RESEND_VERIFY_ACCOUNT_CONTACT_COMPLETED] (state) {
      state.resendDone = true
      state.error = undefined
    },

    [CLEAR_STATES] (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    // DEPRECATED as verification flow in views/Verification.vue is not used anymore
    async contactVerificationInitFromId ({ commit, dispatch, rootState: { client = undefined } }, contactId) {
      console.warn('contactVerificationInitFromId is DEPRECATED which should not be used.')
      try {
        const response = await client.authcoreClient.listContacts()
        commit(VERIFY_ACCOUNT_CONTACT_INIT, {
          contact: response.contacts.filter(item => item.id === contactId.toString()),
          oldContact: true
        })
      } catch (err) {
        commit(VERIFY_ACCOUNT_CONTACT_FAILED, err)
      }
    },
    // DEPRECATED as verification flow in views/Verification.vue is not used anymore
    async resendContactVerification ({ state, commit, rootState: { client = undefined } }, resend = false) {
      console.warn('resendContactVerification is DEPRECATED which should not be used.')
      try {
        commit(VERIFY_ACCOUNT_CONTACT_RESET)
        await client.authcoreClient.startVerifyContact(state.contact.id)
        if (resend) {
          commit(RESEND_VERIFY_ACCOUNT_CONTACT_COMPLETED)
        }
      } catch (err) {
        commit(VERIFY_ACCOUNT_CONTACT_FAILED, err)
      }
    },
    // Used for contact verification in register flow, to list contact from server and record the first one in the store
    async contactVerificationInit ({ commit, rootState: { client = undefined } }) {
      try {
        const response = await client.authcoreClient.listContacts()
        commit(VERIFY_ACCOUNT_CONTACT_INIT, {
          contact: response.contacts,
          oldContact: false
        })
      } catch (err) {
        commit(VERIFY_ACCOUNT_CONTACT_FAILED, err)
      }
    },
    // Used to create authorization token after verification in register flow.
    async createAuthorizationToken ({ commit, rootState: { client = undefined } }, payload) {
      try {
        const { codeChallenge, codeChallengeMethod } = payload
        const response = await client.authcoreClient.createAuthorizationToken(codeChallenge, codeChallengeMethod)
        commit(CREATE_AUTHORIZATION_TOKEN_COMPLETED, response)
      } catch (err) {
        commit(CREATE_AUTHORIZATION_TOKEN_FAILED, err)
      }
    },
    // Used to start primary contact verification flow again, for resend verification in corresponding view.
    async resendPrimaryContactVerification ({ commit, rootState: { client = undefined } }, type) {
      try {
        await client.authcoreClient.startVerifyPrimaryContact(type)
        commit(RESEND_VERIFY_ACCOUNT_CONTACT_COMPLETED)
      } catch (err) {
        commit(VERIFY_ACCOUNT_CONTACT_FAILED, err)
      }
    },
    // Used to send verification code to server in register flow
    async primaryContactVerification ({ state, commit, rootState: { client = undefined } }, { type, verifyCode }) {
      try {
        commit(VERIFY_ACCOUNT_CONTACT_SPINNER_STATE, true)
        commit(VERIFY_ACCOUNT_CONTACT_STARTED)
        await client.authcoreClient.verifyPrimaryContact(type, verifyCode)
        commit(VERIFY_ACCOUNT_CONTACT_COMPLETED)
      } catch (err) {
        commit(VERIFY_ACCOUNT_CONTACT_FAILED, err)
      }
    },
    // Used to send verification code to server in profile flow
    async completePrimaryContactVerification ({ commit, rootState: { client = undefined } }, { type, verifyCode }) {
      try {
        commit(VERIFY_ACCOUNT_CONTACT_SPINNER_STATE, true)
        commit(VERIFY_ACCOUNT_CONTACT_STARTED)
        await client.authcoreClient.completeVerifyPrimaryContact(type, verifyCode)
        commit(VERIFY_ACCOUNT_CONTACT_COMPLETED)
      } catch (err) {
        commit(VERIFY_ACCOUNT_CONTACT_FAILED, err)
      }
    }
  }
}
