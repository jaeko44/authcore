import client from '@/client'

function getDefaultState () {
  return {
    loading: false,
    error: {
      username: undefined,
      contact: undefined
    },
    done: false
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    SET_DONE (state) {
      state.loading = false
      state.done = true
    },
    SET_ERROR (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    // Clear states
    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async create ({ commit }, data) {
      try {
        commit('SET_LOADING')
        const { password, ...user } = data
        user.verifier = await client.utils.createPasswordVerifier(data.password)
        await client.client.createUser(user)
        commit('SET_DONE')
        commit('management/alert/SET_MESSAGE', {
          type: 'success',
          message: 'user_list.message.user_created_successfully',
          pending: true
        }, { root: true })
      } catch (err) {
        let error = err
        let contactError
        let usernameError
        if (err.response !== undefined) {
          switch (err.response.status) {
            case 400:
            case 409: {
              const errorFields = err.response.body.details[0].field_violations.map(field => field.field)
              const errorEmail = errorFields.includes('email')
              const errorPhone = errorFields.includes('phone')
              contactError = errorEmail ? 'error.invalid_email' : undefined
              if (contactError === undefined) {
                contactError = errorPhone ? 'error.invalid_phone' : undefined
              }
              if (errorEmail && errorPhone) {
                if (err.response.status === 400) {
                  contactError = 'error.invalid_contact'
                } else if (err.response.status === 409) {
                  contactError = 'error.duplicate_contact'
                }
              }

              usernameError = undefined
              const errorUsername = err.response.body.details[0].field_violations.find(field => field.field === 'username')
              if (errorUsername !== undefined) {
                usernameError = errorUsername.description === 'duplicated' ? 'error.duplicate_username' : 'error.invalid_username'
              }
              error = {
                contact: contactError,
                username: usernameError
              }
              break
            }
            case 500:
              error = err
              break
            default:
              error = err
              break
          }
        }
        if (error.contact || error.username) {
          commit('SET_ERROR', error)
        }
        // TODO commit global message error if not fields error
      }
    }
  }
}
