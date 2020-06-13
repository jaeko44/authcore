<template functional>
  <b-nav-item
    class="navbar-general"
    :active="props.active"
    :to="props.to"
  >
    <slot></slot>
  </b-nav-item>
</template>

<script>
export default {
  name: 'NavbarItem',
  props: {
    active: {
      type: Boolean,
      default: false
    },
    to: {
      type: Object
    }
  }
}
</script>

<style lang="scss">
.navbar-general {
    .nav-link {
        padding: 0.5rem 1.5rem;
        color: $bsq-grey-medium;

        &.active,
        &.router-link-active {
            color: $primary;
        }
    }
}
</style>
