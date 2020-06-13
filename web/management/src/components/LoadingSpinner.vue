<template>
  <div class="text-center">
    <img
      alt="spinner"
      src="@/assets/Load.svg"
      class="rotate"
      @animationstart="rotateStart"
      @animationiteration="rotateIteration" />
  </div>
</template>

<style>
@keyframes spin { 100% { -webkit-transform: rotate(360deg); transform: rotate(360deg); } }

.rotate {
    animation: spin 1.5s cubic-bezier(0, 0.6, 0.36, 1) infinite;
    animation-delay: 0.4s;
    animation-play-state: running;
}

.rotate-finished {
    animation-play-state: paused !important;
}
</style>

<script>
/** LoadingSpinner serves as loading screen for actions
 * The animation includes a little halt for each cycle. It is intentionally made as the halt serves as entry for the next step. To enter next step at the time frame of the halt, follow below:
 *
 * Listen to the the event @spinning, it will emit the status whether the spinner is spinning or not.
 **/
export default {
  name: 'LoadingSpinner',
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
