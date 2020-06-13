<script>
import { mapState } from 'vuex'

import store from '@/store'

const REFRESH_TOKEN_KEY = 'io.authcore.refreshToken'

export default {
  name: 'RefreshToken',
  props: {},

  data () {
    return {}
  },

  computed: {
    ...mapState('client', [
      'authcoreClient'
    ]),
    ...mapState('widgets/account/loginAccountForm', [
      'createAccessTokenError'
    ])
  },

  watch: {
    async authcoreClient (newVal) {
      if (newVal !== undefined) {
        await this.updateAccessToken()
      }
    }
  },

  created () {},
  async mounted () {
    if (this.authcoreClient !== undefined) {
      await this.updateAccessToken()
    }
  },
  updated () {},
  destroyed () {},

  methods: {
    async updateAccessToken () {
      const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
      if (refreshToken) {
        await store.dispatch('widgets/account/loginAccountForm/createAccessTokenByRefreshToken', refreshToken)
        if (this.createAccessTokenError !== undefined) {
          localStorage.removeItem('io.authcore.refreshToken')
          this.postMessage('AuthCore_onTokenUpdatedFail', {})
        }
        const accessToken = this.authcoreClient.getAccessToken()
        this.postMessage('AuthCore_onTokenUpdated', {
          accessToken: accessToken
        })
      } else {
        this.postMessage('AuthCore_onTokenUpdatedFail', {})
      }
    }
  },

  render (h) {
    return h()
  }
}
</script>
