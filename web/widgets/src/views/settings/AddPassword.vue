<template>
  <modify-password
    :title="$t('add_password.title')"
    :description="$t('add_password.description')"
    :button-text="$t('add_password.button.add_password')"
    :password.sync="password"
    :confirm-password.sync="confirmPassword"
    :loading="loading"
    :done="done"
    :password-error="newPasswordError"
    @submit.prevent="addPassword"
    @score="updatePasswordScore"
  />
</template>

<script>
import { mapState, mapActions } from 'vuex'

import ModifyPassword from '@/components/ModifyPassword.vue'

export default {
  name: 'AddPassword',
  components: {
    ModifyPassword
  },
  props: {},

  data () {
    return {}
  },

  computed: {
    ...mapState('password', [
      'loading',
      'done',
      'error'
    ]),
    password: {
      get () {
        return this.$store.state.password.newPassword
      },
      set (value) {
        this.$store.commit('password/SET_NEW_PASSWORD', value)
      }
    },
    confirmPassword: {
      get () {
        return this.$store.state.password.confirmNewPassword
      },
      set (value) {
        this.$store.commit('password/SET_CONFIRM_NEW_PASSWORD', value)
      }
    },
    passwordScore: {
      get () {
        return this.$store.state.password.passswordScore
      },
      set (value) {
        this.$store.commit('password/SET_PASSWORD_SCORE', value)
      }
    },
    newPasswordError () {
      return this.error && this.error.newPassword
    }
  },

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {
    this.$store.commit('password/RESET')
  },

  methods: {
    ...mapActions('password', [
      'addPassword'
    ]),
    updatePasswordScore (score) {
      this.passwordScore = score
    }
  }
}
</script>
