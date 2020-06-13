<template>
  <b-row class="mt-4">
    <b-col cols="12">
      <b-form @submit.prevent="onSubmit" class="mt-4">
        <b-row>
          <b-col>
            <b-form-contact-input
              v-focus
              v-model="handle"
              :state="error ? false : null"
              aria-describedby="handle-error"
              :label="$t('sign_in.input.label.contact')"
              autocorrect="off"
              autocomplete="off"
              autocapitalize="off"
              :spellcheck="false"
              data-cy="handle"
            />
            <b-form-invalid-feedback id="handle-error">
              <i18n v-if="isInputHandleNotFound" :path="handleErrorNotFoundKey" tag="span">
                <template #link>
                  <b-link class="text-danger font-weight-bold" :to="to">{{ $t('sign_in.input.error.create_account') }}</b-link>
                </template>
              </i18n>
              <span v-else>{{ error || $t('general.blank') }}</span>
            </b-form-invalid-feedback>
          </b-col>
        </b-row>
        <b-row class="mb-4">
          <b-col class="text-center">
            <with-loading-button
              block
              :class="{ 'w-75': buttonSize === 'normal' }"
              class="d-inline-block"
              type="submit"
              variant="primary"
              :loading="loading"
            >
              {{ $t('reset_password.button.send_reset_link') }}
            </with-loading-button>
          </b-col>
        </b-row>
      </b-form>
    </b-col>
  </b-row>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import WithLoadingButton from '@/components/WithLoadingButton.vue'

import { INPUT_HANDLE_NOT_FOUND_ERROR } from '@/store/types'

export default {
  name: 'ResetPasswordPane',
  components: {
    WithLoadingButton
  },

  computed: {
    ...mapState('client', [
      'widgetsSettings'
    ]),
    ...mapState('preferences', [
      'buttonSize'
    ]),
    ...mapState('authn', [
      'authnState',
      'error',
      'loading'
    ]),
    isInputHandleNotFound () {
      return this.error === INPUT_HANDLE_NOT_FOUND_ERROR
    },
    handleErrorNotFoundKey () {
      if (this.widgetsSettings.signUpEnabled) {
        return 'sign_in.input.error.user_not_found_allow_create_account'
      }
      return 'sign_in.input.error.user_not_found_not_allow_create_account'
    },
    handle: {
      get () {
        return this.$store.state.authn.handle
      },
      set (value) {
        this.$store.commit('authn/SET_HANDLE', value)
      }
    }
  },

  methods: {
    ...mapActions('authn', [
      'startPasswordReset'
    ]),

    onSubmit () {
      this.startPasswordReset({ handle: this.handle })
    }
  }
}
</script>
