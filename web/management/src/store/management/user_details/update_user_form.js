import client from '@/client'

function getDefaultState () {
  return {
    user: {},

    loading: false,
    done: false,
    error: {
      username: undefined,
      email: undefined,
      phone: undefined
    }
  }
}

export default {
  namespaced: true,
  modules: {},

  state: getDefaultState(),
  getters: {},

  mutations: {
    // Update user
    SET_LOADING (state) {
      state.loading = true
    },
    SET_USER (state, user) {
      state.user = user
    },
    SET_DONE (state) {
      state.loading = false
      state.done = true
      state.error = getDefaultState().error
    },
    SET_ERROR (state, err) {
      state.loading = false
      state.error = err
    },

    // Clear states
    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },
  actions: {
    async update ({ commit, rootState: { management } }, user) {
      try {
        commit('SET_LOADING')
        const {
          id,
          profileName,
          username,
          email,
          emailVerified,
          phoneNumber,
          phoneNumberVerified
        } = user
        const updatedUser = await client.client.updateUser(id, {
          name: profileName,
          preferred_username: username,
          email: email,
          email_verified: emailVerified,
          phone_number: phoneNumber,
          phone_number_verified: phoneNumberVerified
        })
        commit('management/userDetails/SET_USER', updatedUser, { root: true })
        commit('SET_DONE')
        commit('management/alert/SET_MESSAGE', {
          type: 'success',
          message: 'user_details_profile.message.update_user_successfully',
          pending: false
        }, { root: true })
      } catch (err) {
        console.error(err)
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
        if (error.email || error.phone || error.username) {
          commit('SET_ERROR', error)
        }
        // TODO: Handle general error
      }
    }
  }
}
