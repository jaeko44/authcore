<template>
  <widget-layout-v2
    :title="$t('change_password.title')"
  >
    <b-row>
      <b-col>
        <b-form @submit.prevent="onSubmit">
          <b-form-group class="mb-0" v-if="currentUser.is_password_set">
            <b-bsq-input
              password
              v-model="oldPassword"
              :state="oldPasswordError ? false : null"
              :disabled="loading"
              :label="$t('change_password.input.label.old_password')"
              aria-describedby="oldPasswordInvalidFeedback"
              autocomplete="current-password"
              type="password"
            />
            <b-form-invalid-feedback id="oldPasswordInvalidFeedback">
              {{ oldPasswordError || $t('general.blank') }}
            </b-form-invalid-feedback>
          </b-form-group>
          <b-form-group class="mb-0">
            <b-bsq-input
              password
              v-model="newPassword"
              :state="newPasswordError ? false : null"
              :disabled="loading"
              :label="$t(`change_password.input.label.new_password`)"
              aria-describedby="newPasswordInvalidFeedback"
              autocomplete="new-password"
              type="password"
            />
            <b-form-invalid-feedback id="newPasswordInvalidFeedback">
              {{ newPasswordError || $t('general.blank') }}
            </b-form-invalid-feedback>
            <div v-if="newPassword" class="mb-4">
              <password-strength-indicator
                :password="newPassword"
                @score="$store.commit('password/SET_PASSWORD_SCORE', $event)"
              />
            </div>
          </b-form-group>
          <b-form-group class="mb-0">
            <b-bsq-input
              password
              v-model="passwordConfirmation"
              :state="null"
              :disabled="loading"
              :label="$t(`modify_password.input.label.confirm_password`)"
              aria-describedby="passwordConfirmationInvalidFeedback"
              autocomplete="new-password"
              type="password"
            />
            <b-form-invalid-feedback id="passwordConfirmationInvalidFeedback">
              {{ $t('general.blank') }}
            </b-form-invalid-feedback>
          </b-form-group>
          <b-button
            block
            type="submit"
            variant="primary"
          >
            {{ $t('change_password.button.change_password') }}
          </b-button>
        </b-form>
      </b-col>
    </b-row>
  </widget-layout-v2>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'
import PasswordStrengthIndicator from '@/components/PasswordStrengthIndicator.vue'

export default {
  name: 'ChangePassword',
  components: {
    WidgetLayoutV2,
    PasswordStrengthIndicator
  },
  props: {},

  computed: {
    ...mapState('users', [
      'currentUser'
    ]),
    ...mapState('authn', {
      oldPasswordError: 'error',
      authnLoading: 'loading'
    }),
    ...mapState('password', {
      newPasswordError: 'error',
      passwordLoading: 'loading'
    }),
    loading () {
      return this.authnLoading || this.passwordLoading
    },
    oldPassword: {
      get () {
        return this.$store.state.authn.password
      },
      set (value) {
        this.$store.commit('authn/SET_PASSWORD', value)
      }
    },
    newPassword: {
      get () {
        return this.$store.state.password.password
      },
      set (value) {
        this.$store.commit('password/SET_PASSWORD', value)
      }
    },
    passwordConfirmation: {
      get () {
        return this.$store.state.password.passwordConfirmation
      },
      set (value) {
        this.$store.commit('password/SET_PASSWORD_CONFIRMATION', value)
      }
    }
  },

  mounted () {
    this.$store.dispatch('users/getCurrentUser')
    this.resetAuthn()
    this.resetPassword()
  },

  methods: {
    ...mapMutations('authn', {
      resetAuthn: 'RESET'
    }),
    ...mapMutations('password', {
      resetPassword: 'RESET'
    }),
    async onSubmit (e) {
      if (this.currentUser.is_password_set) {
        await this.$store.dispatch('authn/verifyPasswordStepUp')
        if (this.oldPasswordError) {
          return
        }
      }
      await this.$store.dispatch('password/changePassword')
      if (!this.newPasswordError) {
        this.$router.push({
          name: 'SettingsHome'
        })
      }
    }
  }
}
</script>
