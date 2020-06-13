<template functional>
  <b-button
    :block="props.block"
    :variant="props.variant"
    :class="[{ 'w-75': props.buttonSize === 'normal' }, { 'd-inline-block': true }, data.class, data.staticClass]"
    v-bind="data.attrs"
    v-on="listeners"
  >
    <span v-if="props.loading">
      <span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>
      <span class="sr-only">Loading...</span>
    </span>
    <span v-else>
      <slot />
    </span>
  </b-button>
</template>

<script>
export default {
  name: 'WithLoadingButton',
  props: {
    block: {
      type: Boolean,
      default: false
    },
    variant: {
      type: String,
      default: 'primary'
    },
    buttonSize: {
      type: String,
      default: 'large'
    },
    class: {
      type: [Array, String],
      default: ''
    },
    loading: {
      type: Boolean,
      default: false
    }
  }
}
</script>
