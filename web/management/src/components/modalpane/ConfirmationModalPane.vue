<template>
  <b-bsq-modal
    v-bind:visible="visible"
    v-on:change="$emit('update:visible', $event)"
    :header-title="headerTitle"
    :button-title="buttonTitle"
    :button-variant="buttonVariant"
    @ok="$emit('ok')"
  >
    <div class="mt-3 mb-5">
      {{ description }}
    </div>
  </b-bsq-modal>
</template>

<script>
export default {
  name: 'ConfirmationModalPane',
  model: {
    prop: 'visible',
    event: 'change'
  },
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    headerTitle: {
      type: String,
      default: ''
    },
    buttonTitle: {
      type: String,
      default: ''
    },
    buttonVariant: {
      type: String,
      default: 'primary'
    },
    description: {
      type: String,
      default: ''
    }
  }
}
</script>
