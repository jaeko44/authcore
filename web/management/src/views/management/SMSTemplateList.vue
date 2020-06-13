<template>
  <general-layout>
    <loading-spinner v-if="!authenticated" />
    <b-container v-else class="mt-5">
      <b-row class="mb-5">
        <b-col>
          <h1 class="font-weight-bold">{{ $t('sms_template_list.title') }}</h1>
        </b-col>
      </b-row>
      <b-row no-gutters>
        <b-col cols="12" class="border rounded">
          <b-row class="p:2.5rem">
            <b-col class="mr-auto">
              <b-container>
                <b-row>
                  <h4 class="font-weight-bold mb-0">{{ $t('sms_template_list.sub_title') }}</h4>
                </b-row>
                <b-row class="text-grey">
                  {{ $t('sms_template_list.description') }}
                </b-row>
              </b-container>
            </b-col>
            <b-col cols="3" class="text-right">
              <template>
                <b-dropdown
                  toggle-class="btn-general"
                  right variant="outline-primary"
                  v-if="selectedLanguage"
                  :text="$t(`common.languages.${selectedLanguage}`)"
                >
                  <b-dropdown-item
                    v-for="option in languageOptions"
                    :key="option.value"
                    :value="option.value"
                    @click="selectedLanguage = option.value">
                    {{ option.text }}
                  </b-dropdown-item>
                </b-dropdown>
              </template>
            </b-col>
          </b-row>

          <data-table
            :items="filteredTemplates"
            :fields="fields"
          >
            <template v-slot="slotProps">
              <b-button
                block
                class="btn-general"
                variant="outline-primary"
                @click="editTemplate(slotProps.data)">
                {{ $t('sms_template_list.button.edit') }}
              </b-button>
            </template>
          </data-table>
        </b-col>
      </b-row>
    </b-container>
  </general-layout>
</template>

<script>
import { mapState } from 'vuex'

import snakeCase from 'lodash/snakeCase'

import router from '@/router'

import GeneralLayout from '@/components/GeneralLayout.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import DataTable from '@/components/DataTable.vue'

export default {
  name: 'SMSTemplateList',
  components: {
    GeneralLayout,
    LoadingSpinner,
    DataTable
  },

  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/templatesList', [
      'smsTemplates',
      'template',
      'languages',
      'error'
    ]),
    filteredTemplates () {
      const newTemplates = this.smsTemplates
        // Fix item from this.smsTemplates. Not modifying this.smsTemplates
        // which should be from server.
        .map(item => Object.assign({}, item))
        .map(item => {
          item.key = item.name
          item.description = this.$t(`model.template.type.${snakeCase(item.name)}.description`)
          item.name = this.$t(`model.template.type.${snakeCase(item.name)}.name`)
          if (item.updatedAt) {
            item.lastUpdateString = item.updatedAt
          } else {
            item.lastUpdateString = this.$t('model.template.default')
          }
          return item
        })
      return newTemplates
    },
    languageOptions () {
      return this.languages.map(language => {
        return {
          value: language,
          text: this.$t(`common.languages.${language}`)
        }
      })
    },
    fields () {
      return [
        { key: 'name', label: this.$t('sms_template_list.text.sms_templates'), class: 'w-22 pl:2.5rem' },
        { key: 'lastUpdateString', label: this.$t('model.template.updated_at'), class: 'w-22' },
        { key: 'description', label: this.$t('model.template.description') },
        { key: 'actions', label: '', class: 'text-right w-22 pr:2.5rem' }
      ]
    }
  },

  data () {
    return {
      selectedLanguage: undefined,
      selectedTemplate: undefined
    }
  },

  watch: {
    languages () {
      if (this.languages.length === 0) return
      this.selectedLanguage = this.languages[0]
    },
    async selectedLanguage (language) {
      if (language === undefined) return
      await this.$store.dispatch('management/templatesList/listSMSTemplates', language)
    }
  },

  mounted () {
    this.$store.dispatch('management/templatesList/listAvailableLanguage')
  },
  destroyed () {},

  methods: {
    async editTemplate (template) {
      this.selectedTemplate = template
      await this.$store.dispatch('management/templatesList/getSMSTemplate', {
        language: this.selectedLanguage,
        name: this.selectedTemplate.key
      })
      if (!this.error) {
        router.push({
          name: 'SMSTemplateEdit',
          params: {
            templateName: this.selectedTemplate.key,
            language: this.selectedLanguage
          }
        })
      }
    }
  }
}
</script>

<style scoped>
.code-editor, .code-editor:focus {
  background-color: #000;
  color: #fff;
  font-family: 'PT Mono', monospace;
}
</style>
