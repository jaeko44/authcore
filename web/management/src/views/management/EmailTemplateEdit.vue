<template>
  <general-layout
    show-bar
    :backAction="routeBack"
    :backTitle="$t('email_template_list.title')"
    id="pageTitle"
  >
    <loading-spinner v-if="!authenticated" />
    <b-container v-else class="mt-5">
      <b-row>
        <b-col>
          <h1 class="my-0 font-weight-bold">{{ $t('email_template_edit.title', { name: $t(`model.template.type.${templateNameSnakeCase}.name`).toLowerCase() }) }}</h1>
        </b-col>
      </b-row>
     <b-row class="my-3">
        <b-col>
          <div class="text-grey">
            {{ $t(`model.template.type.${templateNameSnakeCase}.description`) }}
          </div>
        </b-col>
     </b-row>
     <b-row no-gutters class="mt-5">
        <b-col cols="12" class="border rounded">
          <b-row class="p:2.5rem">
            <b-col>
              <b-alert
                v-if="isShowSuccessAlert"
                variant="success"
                show
              >
                {{ alertMessage }}
              </b-alert>

              <b-form @submit.prevent="update">
                <b-row class="mb-4">
                  <b-col md="9">
                    <b-bsq-input
                      separate
                      v-model="updatedTemplate.subject"
                      :label="$t('model.template.subject')"
                      class="mb-5"
                      input-class="font:0.875rem"
                    />
                  </b-col>
                </b-row>
                <b-row class="mb-5">
                  <b-col md="9">
                    <div class="mb-3">{{ $t('model.template.text_template') }}</div>
                    <codemirror
                      v-model="updatedTemplate.textTemplate"
                      :options="cmTextOptions"
                    ></codemirror>
                  </b-col>
                </b-row>
                <b-row class="mb-5">
                  <b-col md="9">
                    <div class="mb-3">{{ $t('model.template.html_template') }}</div>
                    <codemirror
                      ref="htmlTemplate"
                      v-model="updatedTemplate.htmlTemplate"
                      :options="cmHTMLOptions"
                      @ready="onCmReadyHTML"
                    ></codemirror>
                  </b-col>
                </b-row>
                <b-row class="mb-4">
                  <b-col>
                    <b-button
                      class="mr-3 px-5"
                      variant="primary"
                      @click="saveTemplate"
                    >
                    {{ $t('email_template_edit.button.save') }}
                    </b-button>
                    <b-button
                      id="restore-btn"
                      class="px-5"
                      variant="outline-danger"
                      @click="restoreToDefaultTemplate"
                    >
                      {{ $t('email_template_edit.button.reset') }}
                    </b-button>
                  </b-col>
                </b-row>
              </b-form>
              <b-bsq-modal
                v-model="isModalEnabled"
                :header-title="$t('email_template_edit.modal.title')"
                :button-title="$t('email_template_edit.modal.button.confirm')"
                button-variant="danger"
                @ok="modalSubmit"
              >
                <div class="mt-3 mb-5">
                  <div v-if="isModalEnabled">
                    {{ $t('email_template_edit.modal.description') }}
                  </div>
                </div>
              </b-bsq-modal>
            </b-col>
          </b-row>
        </b-col>
      </b-row>
    </b-container>
  </general-layout>
</template>

<script>
import { mapState } from 'vuex'

import snakeCase from 'lodash/snakeCase'
import cloneDeep from 'lodash/cloneDeep'

import router from '@/router'

import GeneralLayout from '@/components/GeneralLayout.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import { codemirror } from 'vue-codemirror'

import 'codemirror/mode/htmlmixed/htmlmixed'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/base16-dark.css'
import 'codemirror/addon/lint/lint.css'
import 'codemirror/addon/lint/lint'
import 'codemirror/addon/lint/html-lint'

export default {
  name: 'EmailTemplateEdit',
  components: {
    GeneralLayout,
    LoadingSpinner,
    codemirror
  },
  props: {
    templateName: {
      type: String,
      required: true
    },
    language: {
      type: String,
      required: true
    }
  },
  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/templatesList', [
      'template',
      'error'
    ]),
    templateNameSnakeCase () {
      return snakeCase(this.templateName)
    },
    isShowSuccessAlert () {
      return this.isShowSuccessCreateAlert || this.isShowSuccessRestoreToDefaultAlert
    },
    alertMessage () {
      if (this.isShowSuccessCreateAlert) {
        return this.$t('email_template_edit.message.success_create')
      } else if (this.isShowSuccessRestoreToDefaultAlert) {
        return this.$t('email_template_edit.message.success_reset')
      }
      return ''
    },
    cmTextOptions () {
      return {
        indentUnit: 4,
        tabSize: 4,
        mode: 'null',
        lineNumbers: true,
        line: true,
        theme: 'base16-dark'
      }
    },
    cmHTMLOptions () {
      return {
        lint: true,
        indentUnit: 4,
        tabSize: 4,
        mode: 'text/html',
        gutters: ['CodeMirror-lint-markers'],
        lineNumbers: true,
        line: true,
        theme: 'base16-dark'
      }
    }
  },

  data () {
    return {
      isShowSuccessCreateAlert: false,
      isShowSuccessRestoreToDefaultAlert: false,
      isModalEnabled: false,

      updatedTemplate: undefined
    }
  },

  watch: {
    template () {
      this.updatedTemplate = cloneDeep(this.template)
    }
  },

  created () {
    this.updatedTemplate = this.template
    if (this.templateName === undefined || this.language === undefined) {
      this.routeBack()
    }
  },
  mounted () {},
  destroyed () {},

  methods: {
    routeBack () {
      router.push({
        name: 'EmailTemplateList'
      })
    },
    onCmReadyHTML () {
      this.$refs.htmlTemplate.codemirror.setSize(null, '30vh')
    },
    scrollTop () {
      this.$scrollTo('#pageTitle', 100, {
        offset: -60
      })
    },
    async saveTemplate () {
      const { subject, htmlTemplate, textTemplate } = this.updatedTemplate
      await this.$store.dispatch('management/templatesList/createEmailTemplate', {
        language: this.language,
        name: this.templateName,
        subject,
        htmlTemplate,
        textTemplate
      })
      if (!this.error) {
        this.isShowSuccessCreateAlert = true
      }
      this.scrollTop()
      // Keep update for the list
      await this.$store.dispatch('management/templatesList/listEmailTemplates', this.language)
    },
    restoreToDefaultTemplate () {
      this.isModalEnabled = true
      // Unfocus the button to prevent scrolling when modal dismisses.
      document.getElementById('restore-btn').blur()
    },
    async modalSubmit () {
      await this.$store.dispatch('management/templatesList/restoreToDefaultEmailTemplate', {
        language: this.language,
        name: this.templateName
      })
      await this.$store.dispatch('management/templatesList/getEmailTemplate', {
        language: this.language,
        name: this.templateName
      })
      if (!this.error) {
        this.$nextTick(() => {
          this.isModalEnabled = false
        })
        this.isShowSuccessRestoreToDefaultAlert = true
      }
      this.scrollTop()
      // Keep update for the list
      await this.$store.dispatch('management/templatesList/listEmailTemplates', this.language)
    }
  }
}
</script>
