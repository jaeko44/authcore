<template>
  <div>
    <b-row v-if="logoEnabled" id="logo">
      <b-col class="text-center">
        <img v-if="logo === 'default'" class="large-logo-size" alt="" :src="require('@/assets/logo.svg')" />
        <img v-else class="large-logo-size" alt="" :src="logo" />
      </b-col>
    </b-row>
    <b-row v-if="companyEnabled" id="company" class="mt-2">
      <b-col class="h5 font-weight-normal text-center px-0 my-0">
        {{ company }}
      </b-col>
    </b-row>
    <div class="header">
      <div v-if="backButtonEnabled">
        <b-row class="back-row" align-v="center">
          <b-col cols="auto" class="col-auto-with-back-button">
            <!-- Only apply py-2 on button to increase clickable zone for back button -->
            <b-button
              variant="link"
              @click="onBackButton"
              class="border-0 p-2 text-decoration-none text-grey-dark font-size-inherit"
            >
              <i class="ac-icon ac-left-arrow my-0"></i>
            </b-button>
          </b-col>
          <b-col class="h5 py-2 my-0 font-weight-bold text-center">
            {{ title }}
          </b-col>
          <!-- Balancing for the center of the company -->
          <b-col cols="auto" class="pl-0">
            <b-button
              variant="link"
              @click="onBackButton"
              class="border-0 p-2 text-decoration-none text-grey-dark invisible font-size-inherit"
            >
              <i class="ac-icon ac-left-arrow my-0"></i>
            </b-button>
          </b-col>
        </b-row>
        <b-row>
          <b-col class="text-center">
            {{ description }}
          </b-col>
        </b-row>
      </div>
      <div v-else>
        <b-row>
          <b-col class="h5 py-2 my-0 font-weight-bold text-center">
            {{ title }}
          </b-col>
        </b-row>
        <b-row>
          <b-col class="text-center">
            {{ description }}
          </b-col>
        </b-row>
      </div>
      <b-row>
        <b-col cols="12">
          <alert-pane />
        </b-col>
      </b-row>
    </div>
  </div>
</template>

<script>
import AlertPane from '@/components/layout/AlertPane.vue'

export default {
  name: 'LargeIconHeader',
  components: {
    AlertPane
  },
  props: {
    logo: {
      type: String
    },
    company: {
      type: String
    },
    title: {
      type: String
    },
    description: {
      type: String
    },
    backButtonEnabled: {
      type: Boolean,
      default: false
    },
    logoEnabled: {
      type: Boolean,
      default: false
    },
    companyEnabled: {
      type: Boolean,
      default: false
    }
  },

  computed: {},

  methods: {
    onBackButton () {
      this.$emit('back-button')
    }
  }
}
</script>

<style scoped lang="scss">
/* Fix the height of title-row to provide consistent layout no matter with back button or not */
.title-row {
    height: 35px;
}
</style>
