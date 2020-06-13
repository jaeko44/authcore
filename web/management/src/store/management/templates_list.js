import { cloneDeep } from 'lodash'

import client from '@/client'
import { marshalTemplate } from '@/utils/marshal'

function getDefaultState () {
  return {
    emailTemplates: [],
    smsTemplates: [],
    template: undefined,
    languages: [],

    loading: false,
    error: undefined
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
      state.error = undefined
    },

    UNSET_LOADING (state) {
      state.loading = false
    },

    SET_ERROR (state, err) {
      console.error(err)
      state.loading = false
      state.error = err
    },

    RESET_EMAIL_TEMPLATES (state) {
      state.emailTemplates = []
    },

    RESET_SMS_TEMPLATES (state) {
      state.smsTemplates = []
    },

    RESET_LANGUAGES (state) {
      state.languages = []
    },

    SET_LANGUAGES (state, languages) {
      state.languages = cloneDeep(languages)
    },

    SET_EMAIL_TEMPLATES (state, { templates }) {
      state.emailTemplates = cloneDeep(templates)
    },

    SET_SMS_TEMPLATES (state, { templates }) {
      state.smsTemplates = cloneDeep(templates)
    },

    SET_TEMPLATE (state, template) {
      state.template = template
    },

    // Clear states
    RESET_STATES (state) {
      Object.assign(state, getDefaultState())
    }
  },

  actions: {
    async listAvailableLanguage ({ commit }) {
      try {
        commit('SET_LOADING')
        commit('RESET_LANGUAGES')
        const lanugages = await client.client.listTemplateLanguages()
        commit('SET_LANGUAGES', lanugages.results)
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        if (err.response.status === 403) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'error.no_permission',
            pending: false
          }, { root: true })
        }
        commit('SET_ERROR', err)
      }
    },
    async listEmailTemplates ({ commit }, language) {
      try {
        commit('SET_LOADING')
        commit('RESET_EMAIL_TEMPLATES')
        const templates = await client.client.listTemplates('email', language)
        commit('SET_EMAIL_TEMPLATES', { templates: templates.results.map(marshalTemplate) })
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        if (err.response.status === 403) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'error.no_permission',
            pending: false
          }, { root: true })
        }
        commit('SET_ERROR', err)
      }
    },
    async listSMSTemplates ({ commit }, language) {
      try {
        commit('SET_LOADING')
        commit('RESET_SMS_TEMPLATES')
        const templates = await client.client.listTemplates('sms', language)
        commit('SET_SMS_TEMPLATES', { templates: templates.results.map(marshalTemplate) })
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        if (err.response.status === 403) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'error.no_permission',
            pending: false
          }, { root: true })
        }
        commit('SET_ERROR', err)
      }
    },
    async getEmailTemplate ({ commit }, { language, name }) {
      try {
        commit('SET_LOADING')
        const emailTemplate = await client.client.getTemplate('email', language, name)
        commit('SET_TEMPLATE', {
          subject: emailTemplate.subject,
          textTemplate: emailTemplate.text,
          htmlTemplate: emailTemplate.html
        })
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        if (err.response.status === 403) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'error.no_permission',
            pending: false
          }, { root: true })
        }
        commit('SET_ERROR', err)
      }
    },
    async getSMSTemplate ({ commit }, { language, name }) {
      try {
        commit('SET_LOADING')
        const smsTemplate = await client.client.getTemplate('sms', language, name)
        commit('SET_TEMPLATE', {
          template: smsTemplate.text
        })
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        if (err.response.status === 403) {
          commit('management/alert/SET_MESSAGE', {
            type: 'danger',
            message: 'error.no_permission',
            pending: false
          }, { root: true })
        }
        commit('SET_ERROR', err)
      }
    },
    async createEmailTemplate ({ commit }, { language, name, subject, htmlTemplate, textTemplate }) {
      try {
        commit('SET_LOADING')
        const newTemplate = {
          subject,
          html: htmlTemplate,
          text: textTemplate
        }
        await client.client.updateTemplate('email', language, name, newTemplate)
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        commit('SET_ERROR', err)
      }
    },
    async createSMSTemplate ({ commit }, { language, name, template }) {
      try {
        commit('SET_LOADING')
        await client.client.updateTemplate('sms', language, name, { text: template })
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        commit('SET_ERROR', err)
      }
    },
    async restoreToDefaultEmailTemplate ({ commit }, { language, name }) {
      try {
        commit('SET_LOADING')
        await client.client.resetTemplate('email', language, name)
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        commit('SET_ERROR', err)
      }
    },
    async restoreToDefaultSMSTemplate ({ commit }, { language, name }) {
      try {
        commit('SET_LOADING')
        await client.client.resetTemplate('sms', language, name)
        commit('UNSET_LOADING')
      } catch (err) {
        console.error(err)
        commit('SET_ERROR', err)
      }
    }
  }
}
