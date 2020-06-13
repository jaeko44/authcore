<template>
  <div class="d-flex justify-content-center loading-spinner">
    <b-spinner
      class="align-self-center"
      variant="primary"
      @animationstart="rotateStart"
      @animationiteration="rotateIteration"
    />
  </div>
</template>

<script>
/** LoadingSpinner serves as loading screen for actions
 * The animation includes a little halt for each cycle. It is intentionally made as the halt serves as entry for the next step. To enter next step at the time frame of the halt, follow below:
 *
 * Listen to the the event @spinning, it will emit the status whether the spinner is spinning or not.
 **/
export default {
  name: 'LoadingSpinner',
  props: {
    height: {
      type: Number,
      default: 0
    }
  },

  computed: {},

  methods: {
    rotateStart () {
      this.$emit('spinning', true)
    },
    rotateIteration () {
      this.$emit('spinning', false)
      setTimeout(() => {
        this.$emit('spinning', true)
      }, 50)
    }
  }
}
</script>

<style lang="scss">
.loading-spinner {
    min-height: 360px;
}
</style>
