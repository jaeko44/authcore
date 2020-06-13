<template>
  <widget-template
    logo-enabled
    :title="currentTitle"
    :back-enabled="false"
  >
    <b-form>
      <b-row class="mb-4">
        <b-col class="text-center">
          {{ $t(errorContent) }}
        </b-col>
      </b-row>
      <b-row>
        <b-col class="text-center">
          <!-- Check whether it is from OAuth first, applied for both desktop and mobile -->
          <b-link v-if="oauthLink" :href="oauthLink">
            {{ $t('reset_password.text.return_home') }}
          </b-link>
          <b-link v-else-if="inMobile" :href="widgetsSettings.redirectFallbackUrl">
            {{ $t('reset_password.text.return_home') }}
          </b-link>
          <b-link v-else :to="{ name: 'SignIn' }">
            {{ $t('reset_password.text.return_home') }}
          </b-link>
        </b-col>
      </b-row>
    </b-form>
  </widget-template>
</template>

<script>
import { mapState } from 'vuex'
import { i18n } from '@/i18n-setup'
import intersection from 'lodash-es/intersection'

import WidgetTemplate from '@/components/WidgetTemplate.vue'

export default {
  name: 'ErrorPage',
  components: {
    WidgetTemplate
  },
  props: {},

  data () {
    return {
      oauthLink: ''
    }
  },

  computed: {
    ...mapState('client', [
      'containerId',
      'inMobile',
      'widgetsSettings'
    ]),
    ...mapState('widgets/errorPage', [
      'errorKey',
      'errorMessage'
    ]),
    serviceResultString () {
      // Sign in factors allowed in server, should align with what server allowed.
      const allowedServiceArray = ['PASSWORD', 'GOOGLE', 'FACEBOOK', 'TWITTER', 'APPLE', 'MATTERS']
      const serviceArray = this.errorMessage.split(' ')
      const filteredService = intersection(allowedServiceArray, serviceArray)
      let resultString
      if (filteredService.length > 0) {
        resultString = filteredService
          .map(item => i18n.t(`error_page.text.${item.toLowerCase()}`))
          .join(i18n.t('general.separator'))
      } else {
        resultString = i18n.t('error_page.text.others')
      }
      return resultString
    },
    currentTitle () {
      return i18n.t('error_page.title', { service: this.serviceResultString })
    },
    errorContent () {
      return i18n.t('error_page.text.description', { service: this.serviceResultString })
    }
  },

  watch: {},

  created () {},
  mounted () {
    this.oauthLink = sessionStorage.getItem('oauth_link')
  },
  updated () {},
  destroyed () {},

  methods: {}
}
</script>
