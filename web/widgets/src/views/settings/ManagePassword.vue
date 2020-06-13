<template>
  <widget-layout-v2
    :title="$t('manage_password.title')"
    :description="$t('manage_password.description')"
  >
    <b-row>
      <b-col>
        <b-form-group class="mb-0">
          <b-bsq-input
            password
            v-model="currentUser.is_password_set"
            :state="passwordError ? false : null"
            :disabled="loading || done"
            :label="$t('manage_password.input.label.password')"
            aria-describedby="passwordInvalidFeedback"
            autocomplete="current-password"
            type="password"
          />
          <b-form-invalid-feedback id="passwordInvalidFeedback">
            {{ passwordError || $t('general.blank') }}
          </b-form-invalid-feedback>
        </b-form-group>
        <b-button
          block
          type="button"
          class="mb-4"
          variant="outline-primary"
          @click="changePassword"
        >
          {{ $t('manage_password.button.change_password') }}
        </b-button>
        <b-button
          block
          type="button"
          variant="outline-danger"
          @click="removePassword"
        >
          {{ $t('manage_password.button.remove_password') }}
        </b-button>
      </b-col>
    </b-row>
  </widget-layout-v2>
</template>

<script>
import { mapState, mapActions } from 'vuex'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'

import router from '@/router'

export default {
  name: 'ManagePassword',
  components: {
    WidgetLayoutV2
  },
  props: {},

  data () {
    return {}
  },

  computed: {
    ...mapState('password', [
      'loading',
      'done'
    ]),
    ...mapState('users', [
      'currentUser'
    ]),
    password: {
      get () {
        return this.$store.state.password.oldPassword
      },
      set (value) {
        this.$store.commit('password/SET_OLD_PASSWORD', value)
      }
    },
    passwordError () {
      return this.error && this.error.oldPassword
    }
  },

  mounted () {
    this.$store.dispatch('users/getCurrentUser')
  },

  methods: {
    ...mapActions('password', [
      'verifyPassword'
    ]),
    async changePassword () {
      await this.verifyPassword()
      // TODO: Maybe need to check if not error
      router.push({ name: 'ChangePassword' })
    },
    async removePassword () {
      await this.verifyPassword()
      // TODO: Maybe need to check if not error
      router.push({ name: 'RemovePassword' })
    }
  }
}
</script>
