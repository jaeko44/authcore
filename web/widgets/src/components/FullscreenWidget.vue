<template>
  <div class="fullscreen-center-container">
    <div class="fullscreen-center">
      <div class="fullscreen-overlay"></div>
      <div class="fullscreen-center-wrapper">
        <slot name="default"></slot>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
$wrapper-width: 480px;
$wrapper-padding: 30px;

.fullscreen-center-container {
    position: static;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    min-height: 100vh;
}

.fullscreen-center {
    @media only screen and (min-width: $wrapper-width) {
        box-shadow: 1px 1px 4px #cbcbcb;
        margin: auto;
    }
    @media only screen and (max-width: $wrapper-width) {
        width: 100vw;
        margin-left: auto;
        margin-right: auto;
    }
    background: white;
    width: $wrapper-width;
    border-radius: 5px;
}

.fullscreen-overlay {
    @media only screen and (max-width: $wrapper-width) {
        background: white;
    }
    background: #f7f7f7;
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    z-index: -1;
    display: flex;
}

.fullscreen-center-wrapper {
    padding-bottom: $wrapper-padding;
}

</style>

<script>
export default {
  name: 'FullscreenWidget',
  components: {},

  props: {},

  data () {
    return {}
  },

  computed: {},

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {},

  methods: {}
}
</script>
