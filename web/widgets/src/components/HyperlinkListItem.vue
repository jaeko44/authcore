<template>
  <b-list-group-item
    class="generic-list-item"
    :href="href"
    :to="to"
    @click="click"
  >
    <b-row align-v="end">
      <b-col class="pr-0">
        <h5 class="mb-0"><slot name="title">{{ title }}</slot></h5>
      </b-col>
      <b-col class="text-right" cols="auto">
        <i class="arrow ac-icon ac-chevron-right"></i>
      </b-col>
    </b-row>
    <b-row>
      <b-col>
        <slot name="default"></slot>
      </b-col>
      <b-col cols="auto"></b-col>
    </b-row>
  </b-list-group-item>
</template>

<style>
.arrow {
    font-size: 1.1rem;
}
</style>

<script>
/* HyperlinkListItem provides half of grid gutter width as x padding to match the alignment of design. To use the component the parent row/column has to be set without x padding (px-0) */
export default {
  name: 'HyperlinkListItem',
  components: {},

  props: {
    title: {
      type: String,
      default: ''
    },
    href: {
      type: String,
      default: '#'
    },
    to: {
      type: [String, Object],
      default: ''
    }
  },

  data () {
    return {}
  },

  computed: {},

  watch: {},

  created () {},
  mounted () {},
  updated () {},
  destroyed () {},

  methods: {
    click () {
      this.$emit('click')
    }
  }
}
</script>
