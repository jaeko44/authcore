import client from '@/client'

import { marshalRole } from '@/utils/marshal'
import { i18n } from '@/i18n-setup'

function getDefaultState () {
  return {
    loading: false,
    done: false,
    error: undefined,

    roleAssignments: []
  }
}

export default {
  namespaced: true,

  state: getDefaultState(),

  mutations: {
    SET_LOADING (state) {
      state.loading = true
    },
    SET_ROLE_ASSIGNMENTS (state, roleAssignments) {
      state.loading = false
      state.roleAssignments = roleAssignments.map(marshalRole)
    },
    SET_ERROR (state, error) {
      console.error(error)
      state.error = error
    },
    UNSET_ERROR (state) {
      state.error = undefined
    },

    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async fetchRoleAssignmentsList ({ commit }, userId) {
      try {
        commit('SET_LOADING')
        const roleAssignments = await client.client.getUserRoles(userId)
        commit('SET_ROLE_ASSIGNMENTS', roleAssignments.results || [])
      } catch (err) {
        commit('SET_ERROR', err)
      }
    },
    async assignRole ({ commit, dispatch, rootState: { management } }, roleAssignmentId) {
      try {
        commit('UNSET_ERROR')
        commit('SET_LOADING')
        const { id: userId } = management.userDetails.user
        await client.client.assignUserRole(userId, roleAssignmentId)
        await dispatch('fetchRoleAssignmentsList', userId)
      } catch (err) {
        let error = err
        if (err.response) {
          if (err.response.status === 409) {
            commit('management/alert/SET_MESSAGE', {
              type: 'danger',
              message: 'error.role_assignment_exists_for_same_account',
              pending: false
            }, { root: true })
            error = i18n.t('error.assignment_exists')
          }
        }
        commit('SET_ERROR', error)
      }
    },
    async unassignRole ({ commit, dispatch, rootState: { currentUser, management } }, roleAssignmentId) {
      try {
        commit('UNSET_ERROR')
        commit('SET_LOADING')
        const { id: currentUserId } = currentUser.user
        const { id: userId } = management.userDetails.user
        if (currentUserId === userId) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'error.role_unassignment_for_same_account',
            pending: false
          }, { root: true })
          throw new Error('error.role_unassignment_for_same_account')
        }
        await client.client.unassignUserRole(userId, roleAssignmentId)
        await dispatch('fetchRoleAssignmentsList', userId)
      } catch (err) {
        commit('SET_ERROR', err)
      }
    }
  }
}
