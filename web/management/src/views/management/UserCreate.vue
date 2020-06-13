<template>
  <general-layout
    show-bar
    :back-title="$t('user_create.text.back_title')"
    :back-action="routeBack"
  >
    <b-container class="mt-5">
      <b-row class="mb-5">
        <b-col>
          <h1 class="my-0 font-weight-bold">{{ $t('user_create.title') }}</h1>
        </b-col>
      </b-row>
      <b-form @submit.prevent="createUser">
        <b-bsq-input
          separate
          v-model="displayName"
          :disabled="loading"
          :label="$t('model.user.name')"
          class="my-5 w-60"
        />
        <b-bsq-input
          separate
          v-model="username"
          :state="usernameErrorState"
          aria-describedby="username"
          :disabled="loading"
          :label="$t('model.user.username')"
          class="my-5 w-60"
          autocomplete="username"
        >
          <b-form-invalid-feedback id="username">
            {{ $t(usernameErrorContent) }}
          </b-form-invalid-feedback>
        </b-bsq-input>
        <!-- TODO: Email, phone or contact field should be in condition to the server setting -->
        <b-form-contact-input
          separate
          v-model="contact"
          :state="contactErrorState"
          aria-describedby="contact"
          :disabled="loading"
          :label="$t('model.user.email_or_phone')"
          class="my-5 w-60"
        >
          <b-form-invalid-feedback id="contact">
            {{ $t(contactErrorContent) }}
          </b-form-invalid-feedback>
        </b-form-contact-input>
        <b-bsq-input
          separate
          type="password"
          v-model="password"
          :disabled="loading"
          :label="$t('model.user.password')"
          class="my-5 w-60"
        />
        <!-- <h3 class="font-weight-light border-bottom pb-3">{{ $t('user_create.text.generate_password') }}</h3>
             <div class="border w-60 d-flex p-4 my-5">
             <textarea
             id="password"
             :value="generatedPassword"
             class="pt-3 text-center border-0"
             style="resize: none; text-overflow: ellipsis; white-space: pre; overflow-x: hidden;"
             readonly
             />
             <b-button
             variant="outline-primary"
             class="ml-auto align-self-center"
             @click="generatePassword"
             >
             {{ $t('user_create.button.generate') }}
             </b-button> -->
        <!-- TODO: Provide animation -->
        <!-- <b-button
             variant="outline-primary"
             class="ml-5 align-self-center"
             @click="copyToClipboard"
             >
             {{ $t('user_create.button.copy') }}
             </b-button>
             </div> -->
        <b-button
          :disabled="loading"
          type="submit"
          variant="primary"
          class="px-5"
        >
          {{ $t('user_create.button.create_user') }}
        </b-button>
      </b-form>
    </b-container>
  </general-layout>
</template>

<script>
import { mapState } from 'vuex'

import GeneralLayout from '@/components/GeneralLayout.vue'
import { isPossibleNumber } from 'libphonenumber-js'
import { isAdmin } from '@/utils/permission'

export default {
  name: 'UserCreate',
  components: {
    GeneralLayout
  },
  data () {
    return {
      username: '',
      displayName: '',
      contact: '',
      password: ''
    }
  },
  computed: {
    ...mapState('currentUser', [
      'user'
    ]),
    ...mapState('management/usersList/createUserForm', [
      'loading',
      'error',
      'done'
    ]),
    usernameErrorState () {
      return this.error.username === undefined ? null : false
    },
    usernameErrorContent () {
      return this.error.username === undefined ? 'common.blank' : this.error.username
    },
    contactErrorState () {
      return this.error.contact === undefined ? null : false
    },
    contactErrorContent () {
      return this.error.contact === undefined ? 'common.blank' : this.error.contact
    }
  },

  watch: {
    done () {
      // Go back to User list page when create user is done
      this.$router.push({
        name: 'UserList'
      })
    }
  },

  mounted () {
    if (!isAdmin(this.user)) {
      this.$store.commit('management/alert/SET_MESSAGE', {
        type: 'danger',
        message: 'error.no_permission',
        pending: false
      })
    }
  },

  destroyed () {
    this.$store.commit('management/usersList/createUserForm/RESET_STATES')
  },

  methods: {
    routeBack () {
      this.$router.push({
        name: 'UserList'
      })
    },
    async createUser (e) {
      const { username, contact, password, displayName } = this
      const userInformation = {
        username: username,
        password: password,
        displayName: displayName
      }
      if (isPossibleNumber(contact)) {
        userInformation.phone = contact
      } else {
        userInformation.email = contact
      }
      await this.$store.dispatch('management/usersList/createUserForm/create', userInformation)
    }
    /* generatePassword () {
     *   const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOP1234567890'
     *   const length = 12
     *   let password = ""
     *   for (let x = 0; x < length; x++) {
     *     let i = Math.floor(Math.random() * chars.length)
     *     password += chars.charAt(i)
     *   }
     *   this.generatedPassword = password
     * },
     * copyToClipboard () {
     *   const secret = document.getElementById('password')
     *   secret.select()
     *   document.execCommand('copy')
     * } */
  }
}
</script>
