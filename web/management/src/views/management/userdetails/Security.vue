<template>
  <div>
    <password-pane
      :user="user"
    />
    <second-factor-table-pane
      :user="user"
    />
    <oauth-factor-table-pane
      :user="user"
    />
  </div>
</template>
<script>
// Security page for user details.

// For change password action, using action in userDetails store which is the same as user details page to ensure the communciation between backend are the same

import PasswordPane from '@/views/management/userdetails/security/PasswordPane.vue'
import SecondFactorTablePane from '@/views/management/userdetails/security/SecondFactorTablePane.vue'
import OauthFactorTablePane from '@/views/management/userdetails/security/OauthFactorTablePane.vue'

export default {
  name: 'UserDetailsSecurity',
  components: {
    PasswordPane,
    SecondFactorTablePane,
    OauthFactorTablePane
  },

  props: {
    user: {
      type: Object,
      required: true
    }
  }
}
</script>
