<template>
  <b-card
    no-body
    header-tag="header"
    header-bg-variant="white"
    class="h-100"
  >
    <template v-slot:header class="my-0 font-weight-light">
      <h4 class="my-0 font-weight-lighter font:1.375rem">{{ title }}</h4>
    </template>
    <!-- Click zone is defined by using utilities 'stretched-link', check Bootstrap doc for details -->
    <slot v-if="customClickzone" name="default"></slot>
    <b-card-body v-if="!customClickzone" class="position-relative">
      <b-card-text class="font:0.75rem">
        <b-link class="stretched-link" :to="to" :href="href"></b-link>
        <slot name="default"></slot>
      </b-card-text>
    </b-card-body>
  </b-card>
</template>

<script>
export default {
  name: 'InfoCard',
  components: {},

  props: {
    title: {
      type: String,
      default: ''
    },
    customClickzone: {
      type: Boolean,
      default: false
    },
    href: {
      type: String,
      default: null
    },
    to: {
      type: [String, Object],
      default: null
    }
  }
}
</script>
