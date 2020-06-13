<template>
  <div>
    <b-row no-gutters class="mt-5">
      <b-col cols="12" class="border rounded">
        <b-row class="p:2.5rem">
          <b-col>
            <h4 class="font-weight-bold">{{ $t('user_details_metadata.title.user_metadata') }}</h4>
            <div class="text-grey">{{ $t('user_details_metadata.description.user_metadata') }}</div>
            <b-row class="my-5">
              <b-col cols="9">
                <codemirror
                  v-model="formattedUserMetadata"
                  :options="cmOptions"
                ></codemirror>
              </b-col>
            </b-row>
            <b-row>
              <b-col md="3">
                <b-button block class="btn-general" variant="primary" @click="updateUserMetadata"
                >
                  {{ $t('user_details_metadata.button.save') }}
                </b-button>
              </b-col>
            </b-row>
          </b-col>
        </b-row>
      </b-col>
    </b-row>

    <b-row no-gutters class="mt-5">
      <b-col cols="12" class="border rounded">
        <b-row class="p:2.5rem">
          <b-col>
            <h4 class="font-weight-bold">{{ $t('user_details_metadata.title.app_metadata') }}</h4>
            <div class="text-grey">{{ $t('user_details_metadata.description.app_metadata') }}</div>
            <b-row class="my-5">
              <b-col cols="9">
                <codemirror
                  v-model="formattedAppMetadata"
                  :options="cmOptions"
                ></codemirror>
              </b-col>
            </b-row>
            <b-row>
              <b-col md="3">
                <b-button block class="btn-general" variant="primary" @click="updateAppMetadata">
                  {{ $t('user_details_metadata.button.save') }}
                </b-button>
              </b-col>
            </b-row>
          </b-col>
        </b-row>
      </b-col>
    </b-row>
  </div>
</template>

<script>
// Set the json lint for codemirror
import jsonlint from 'jsonlint-mod'

import { mapState } from 'vuex'
import { codemirror } from 'vue-codemirror'
import { js_beautify as jsBeautify } from 'js-beautify'

import 'codemirror/mode/javascript/javascript'
import 'codemirror/lib/codemirror.css'
import 'codemirror/theme/base16-dark.css'
import 'codemirror/addon/lint/lint.css'
import 'codemirror/addon/lint/lint'
import 'codemirror/addon/lint/json-lint'

window.jsonlint = jsonlint

const emptyJSON = '{\n\t\n}'

export default {
  name: 'UserDetailsMetadata',
  components: {
    codemirror
  },

  props: {
    id: {
      type: Number,
      required: true
    }
  },

  data () {
    return {
      updatedUserMetadata: '',
      updatedAppMetadata: ''
    }
  },

  computed: {
    ...mapState('currentUser', [
      'authenticated'
    ]),
    ...mapState('management/userDetails', [
      'user'
    ]),
    ...mapState('management/userDetails/metadata', [
      'loading',
      // only use the update signal to show alerts
      'updateDone',
      'updateError',

      'userMetadata',
      'appMetadata'
    ]),
    isLoadingStatus () {
      return !this.authenticated || this.loading
    },
    cmOptions () {
      return {
        lint: true,
        indentUnit: 4,
        tabSize: 4,
        mode: {
          name: 'javascript',
          json: true
        },
        gutters: ['CodeMirror-lint-markers'],
        lineNumbers: true,
        line: true,
        theme: 'base16-dark'
      }
    },
    formattedUserMetadata: {
      get () {
        if (!this.userMetadata) {
          return emptyJSON
        }
        return jsBeautify(JSON.stringify(this.userMetadata))
      },
      set (newValue) {
        this.updatedUserMetadata = newValue
      }
    },
    formattedAppMetadata: {
      get () {
        if (!this.appMetadata) {
          return emptyJSON
        }
        return jsBeautify(JSON.stringify(this.appMetadata))
      },
      set (newValue) {
        this.updatedAppMetadata = newValue
      }
    }
  },

  async mounted () {
    // get user when not enter via user device
    if (!this.user.id) await this.$store.dispatch('management/userDetails/get', this.id)
    await this.$store.dispatch('management/userDetails/metadata/list', this.user)
  },

  destroyed () {
    this.$store.commit('management/userDetails/metadata/RESET_STATES')
  },

  methods: {
    async updateUserMetadata () {
      await this.$store.dispatch('management/userDetails/metadata/updateUserMetadata', {
        id: this.id,
        userMetadata: this.updatedUserMetadata
      })
      this.$scrollTo('body')
    },
    async updateAppMetadata () {
      await this.$store.dispatch('management/userDetails/metadata/updateAppMetadata', {
        id: this.id,
        appMetadata: this.updatedAppMetadata
      })
      this.$scrollTo('body')
    }
  }
}
</script>
