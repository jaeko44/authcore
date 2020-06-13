<template>
  <widget-layout-v2
    :title="title"
    :description="description"
    @back-button="onBackButton"
  >
    <b-row>
      <b-col>
        <b-button
          block
          type="button"
          variant="danger"
          @click="$emit('click')"
        >
          {{ buttonText }}
        </b-button>
      </b-col>
    </b-row>
  </widget-layout-v2>
</template>

<script>
import WidgetLayoutV2 from '@/components/WidgetLayoutV2.vue'

export default {
  name: 'ConfirmationLayout',
  components: {
    WidgetLayoutV2
  },
  props: {
    title: {
      type: String,
      default: ''
    },
    description: {
      type: String,
      default: ''
    },
    buttonText: {
      type: String,
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
    onBackButton () {
      this.$emit('back-button')
    }
  }
}
</script>
