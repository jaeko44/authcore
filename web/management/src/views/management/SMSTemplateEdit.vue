<template>
  <general-layout
    show-bar
    :backAction="routeBack"
    :backTitle="$t('sms_template_list.title')"
    id="pageTitle"
  >
    <loading-spinner v-if="!authenticated" />
    <b-container v-else class="mt-5">
      <b-row>
        <b-col>
          <h1 class="my-0 font-weight-bold">{{ $t('sms_template_edit.title', { name: $t(`model.template.type.${templateNameSnakeCase}.name`).toLowerCase() }) }}</h1>
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
                    <div class="mb-3">{{ $t('model.template.text_template') }}</div>
                    <codemirror
                      v-model="updatedTemplate.template"
                      :options="cmOptions"
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
                    {{ $t('sms_template_edit.button.save') }}
                    </b-button>
                    <b-button
                      id="restore-btn"
                      class="px-5"
                      variant="outline-danger"
                      @click="restoreToDefaultTemplate"
                    >
                      {{ $t('sms_template_edit.button.reset') }}
                    </b-button>
                  </b-col>
                </b-row>
              </b-form>

              <b-bsq-modal
                v-model="isModalEnabled"
                :header-title="$t('sms_template_edit.modal.title')"
                :button-title="$t('sms_template_edit.modal.button.confirm')"
                button-variant="danger"
                @ok="modalSubmit"
              >
                <div class="mt-3 mb-5">
                  <div v-if="isModalEnabled">
                    {{ $t('sms_template_edit.modal.description') }}
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

import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/base16-dark.css'

export default {
  name: 'SMSTemplateEdit',
  components: {
    codemirror,
    GeneralLayout,
    LoadingSpinner
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
        return this.$t('sms_template_edit.message.success_create')
      } else if (this.isShowSuccessRestoreToDefaultAlert) {
        return this.$t('sms_template_edit.message.success_reset')
      }
      return ''
    },
    cmOptions () {
      return {
        indentUnit: 4,
        tabSize: 4,
        mode: 'null',
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
        name: 'SMSTemplateList'
      })
    },
    scrollTop () {
      this.$scrollTo('#pageTitle', 100, {
        offset: -60
      })
    },
    async saveTemplate () {
      const { template } = this.updatedTemplate
      await this.$store.dispatch('management/templatesList/createSMSTemplate', {
        language: this.language,
        name: this.templateName,
        template
      })
      if (!this.actionError) {
        this.isShowSuccessCreateAlert = true
      }
      this.scrollTop()
      // Keep update for the list
      await this.$store.dispatch('management/templatesList/listSMSTemplates', this.language)
    },
    restoreToDefaultTemplate () {
      this.isModalEnabled = true
      // Unfocus the button to prevent scrolling when modal dismisses.
      document.getElementById('restore-btn').blur()
    },
    async modalSubmit () {
      await this.$store.dispatch('management/templatesList/restoreToDefaultSMSTemplate', {
        language: this.language,
        name: this.templateName
      })
      await this.$store.dispatch('management/templatesList/getSMSTemplate', {
        language: this.language,
        name: this.templateName
      })
      if (!this.actionError) {
        this.$nextTick(() => {
          this.isModalEnabled = false
        })
        this.isShowSuccessRestoreToDefaultAlert = true
      }
      this.scrollTop()
      // Keep update for the list
      await this.$store.dispatch('management/templatesList/listSMSTemplates', this.language)
    }
  }
}
</script>
