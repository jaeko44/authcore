<template>
  <confirmation-layout
    :back-enabled="backEnabled"
    :title="$t('social_login_delete.title')"
    :description="description"
    :button-text="$t('social_login_delete.button.disconnect')"
    @click="removeSocialPlatform"
    @back-button="$router.push({ name: 'SocialLogins' })"
  />
</template>

<script>
import { mapState, mapGetters, mapActions } from 'vuex'

import ConfirmationLayout from '@/components/ConfirmationLayout.vue'

export default {
  name: 'SocialLoginDelete',
  components: {
    ConfirmationLayout
  },

  props: {
    id: {
      type: String,
      required: true
    }
  },

  data () {
    return {}
  },

  computed: {
    ...mapGetters('socialLogin', [
      'getOAuthFactorById'
    ]),
    ...mapState('socialLogin', [
      'loading',
      'done',
      'error',
      'oauthFactors'
    ]),
    selectedSocialLogin () {
      const id = this.id
      return this.oauthFactors.find(oauthFactor => oauthFactor.service === id)
    },
    socialPlatform () {
      return this.selectedSocialLogin ? this.selectedSocialLogin.service.toLowerCase() : ''
    },
    backEnabled () {
      return true
    },
    description () {
      return this.socialPlatform !== '' ? this.$t(`social_login_delete.text.${this.socialPlatform}`) : ''
    }
  },

  destroyed () {
    this.$store.commit('socialLogin/list')
  },

  methods: {
    ...mapActions('socialLogin', [
      'delete'
    ]),
    removeSocialPlatform () {
      this.delete(this.id)
      this.$router.push({ name: 'SocialLogins' })
    }
  }
}
</script>
