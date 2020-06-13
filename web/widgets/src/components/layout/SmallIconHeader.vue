<template>
  <div>
    <b-row
      v-if="backButtonEnabled"
      class="mt-n4 py-2 border-bottom"
      align-v="center"
    >
      <b-col
        cols="auto"
        class="pl-3 pr-2"
        style="margin-left: 10px;"
      >
        <!-- Only apply py-2 on button to increase clickable zone for back button -->
        <b-button
          variant="link"
          @click="onBackButton"
          class="border-0 p-2 text-decoration-none text-grey-dark font-size-inherit"
        >
          <i class="ac-icon ac-left-arrow my-0 font:1rem"></i>
        </b-button>
      </b-col>
      <div v-if="logoEnabled">
        <img v-if="logo === 'default'" class="small-logo-size" alt="" :src="require('@/assets/logo.svg')" />
        <img v-else class="small-logo-size" alt="" :src="logo" />
      </div>
      <div v-if="companyEnabled" class="my-0 pl-3 font:1rem">
        {{ company }}
      </div>
    </b-row>
    <b-row
      v-else
      class="mt-n4 py-2 border-bottom"
    >
      <b-col :class="{ 'py-1': logoEnabled, 'py-2': !logoEnabled }">
        <span v-if="logoEnabled">
          <img v-if="logo === 'default'" class="small-logo-size" alt="" :src="require('@/assets/logo.svg')" />
          <img v-else class="small-logo-size" alt="" :src="logo" />
        </span>
        <span class="my-0 pl-3 font:1rem py-2">
          {{ company }}
        </span>
      </b-col>
    </b-row>
    <b-row>
      <b-col cols="12">
        <alert-pane />
      </b-col>
    </b-row>
    <!-- Provide margin for mobile case as title could be multiple line -->
    <div id="title">
      <b-row class="mt-4">
        <b-col class="h5 my-0 font-weight-bold">
          {{ title }}
        </b-col>
      </b-row>
      <b-row class="mb-4">
        <b-col v-if="description">
          {{ description }}
        </b-col>
        <b-col v-else>
          <slot name="description"></slot>
        </b-col>
      </b-row>
    </div>
  </div>
</template>

<script>
import AlertPane from '@/components/layout/AlertPane.vue'

export default {
  name: 'SmallIconHeader',
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
      default: true
    }
  },

  computed: {
    logoEnabled () {
      return this.logo
    },
    companyEnabled () {
      return this.company
    }
  },

  methods: {
    onBackButton () {
      this.$emit('back-button')
    }
  }
}
</script>
