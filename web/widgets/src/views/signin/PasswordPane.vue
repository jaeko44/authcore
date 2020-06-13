<template>
  <b-row>
    <b-col cols="12">
      <b-form @submit.prevent="verifyPassword(password)">
        <b-row class="my-4" align-h="center">
          <b-col class="h5 my-0 text-grey text-center" cols="12">
            {{ handle }}
          </b-col>
        </b-row>
        <b-row>
          <b-col>
            <input class="d-none" name="handle" type="text" :value="handle" />
            <b-bsq-input
              password
              v-focus
              v-model="password"
              :state="error ? false : null"
              aria-describedby="code-error"
              :label="$t('sign_in.input.label.password')"
              type="password"
            />
            <b-form-invalid-feedback id="code-error">
              {{ error || $t('general.blank') }}
            </b-form-invalid-feedback>
          </b-col>
        </b-row>
        <b-row class="mb-4">
          <b-col class="text-center">
            <with-loading-button
              block
              type="submit"
              :button-size="buttonSize"
              :loading="loading"
            >
              {{ $t('sign_in.button.next') }}
            </with-loading-button>
          </b-col>
        </b-row>
      </b-form>
      <b-row>
        <b-col class="text-center">
          <b-link
            class="font-weight-bold"
            @click="$router.push({name: 'ResetPassword', query: { handle }})"
          >
            {{ $t('sign_in.link.forgot_password') }}
          </b-link>
        </b-col>
      </b-row>
    </b-col>
  </b-row>
</template>

<script>
import { mapState, mapActions } from 'vuex'

import WithLoadingButton from '@/components/WithLoadingButton.vue'

export default {
  name: 'PasswordPane',

  components: {
    WithLoadingButton
  },

  computed: {
    ...mapState('preferences', [
      'buttonSize'
    ]),
    ...mapState('authn', [
      'error',
      'loading',
      'handle'
    ]),
    password: {
      get () {
        return this.$store.state.authn.password
      },
      set (value) {
        this.$store.commit('authn/SET_PASSWORD', value)
      }
    }
  },

  methods: {
    ...mapActions('authn', [
      'verifyPassword'
    ])
  }
}
</script>
