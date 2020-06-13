<template>
  <widget-template
    v-if="resetPasswordDone"
    logo-enabled
    :back-enabled="false"
    :logo="logo"
    :title="currentTitle"
  >
    <div>
      <b-row class="mb-3">
        <b-col class="text-center">
          {{ $t('reset_password.text.reset_password_success') }}
        </b-col>
      </b-row>
      <b-row>
        <b-col>
          <b-button
            v-if="redirectUri"
            block
            :href="decodeURIComponent(redirectUri)"
            variant="primary"
          >
            {{ $t('reset_password.button.sign_in') }}
          </b-button>
          <b-button
            v-else
            block
            variant="primary"
            @click="redirectToSignIn"
          >
            {{ $t('reset_password.button.sign_in') }}
          </b-button>
        </b-col>
      </b-row>
    </div>
  </widget-template>
</template>

<script>
import { mapState } from 'vuex'

import router from '@/router'
import store from '@/store'
import { i18n } from '@/i18n-setup'

import WidgetTemplate from '@/components/WidgetTemplate.vue'

export default {
  name: 'ResetPasswordCompleted',
  components: {
    WidgetTemplate
  },
  props: {
    company: {
      type: String
    },
    logo: {
      type: String
    },
    redirectUri: {
      type: String
    }
  },

  data () {
    return {}
  },

  computed: {
    ...mapState('widgets/account/resetPasswordForm', [
      'resetPasswordDone'
    ]),
    currentTitle () {
      return i18n.t('reset_password.title')
    }
  },

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {
    store.commit('widgets/account/resetPasswordForm/CLEAR_STATES')
  },

  methods: {
    redirectToSignIn () {
      router.push({ name: 'SignIn' })
    }
  }
}
</script>
