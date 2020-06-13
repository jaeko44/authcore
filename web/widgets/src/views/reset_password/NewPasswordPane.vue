<template>
  <b-row class="mt-4">
    <b-col cols="12">
      <b-form @submit.prevent="onSubmit">
        <b-form-group class="mb-0">
          <b-bsq-input
            password
            v-model="newPassword"
            :state="error ? false : null"
            :disabled="loading"
            :label="$t(`change_password.input.label.new_password`)"
            aria-describedby="newPasswordInvalidFeedback"
            autocomplete="new-password"
            type="password"
          />
          <b-form-invalid-feedback id="newPasswordInvalidFeedback">
            {{ error || $t('general.blank') }}
          </b-form-invalid-feedback>
          <div v-if="newPassword" class="mb-4">
            <password-strength-indicator
              :password="newPassword"
              @score="$store.commit('authn/SET_PASSWORD_SCORE', $event)"
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
          :disabled="!newPassword || !passwordConfirmation"
        >
          {{ $t('change_password.button.change_password') }}
        </b-button>
      </b-form>
    </b-col>
  </b-row>
</template>

<script>
import { mapState, mapActions } from 'vuex'

import PasswordStrengthIndicator from '@/components/PasswordStrengthIndicator.vue'

export default {
  name: 'NewPasswordPane',
  components: {
    PasswordStrengthIndicator
  },

  computed: {
    ...mapState('preferences', [
      'buttonSize'
    ]),
    ...mapState('authn', [
      'authnState',
      'loading',
      'error',
      'handle',
      'passwordScore'
    ]),
    newPassword: {
      get () {
        return this.$store.state.authn.password
      },
      set (value) {
        this.$store.commit('authn/SET_PASSWORD', value)
      }
    },
    passwordConfirmation: {
      get () {
        return this.$store.state.authn.passwordConfirmation
      },
      set (value) {
        this.$store.commit('authn/SET_PASSWORD_CONFIRMATION', value)
      }
    }
  },

  methods: {
    ...mapActions('authn', [
      'verifyPasswordReset'
    ]),

    onSubmit () {
      this.verifyPasswordReset({ stateToken: this.authnState.state_token })
    }
  }
}
</script>
