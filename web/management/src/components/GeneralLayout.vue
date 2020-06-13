<!-- GeneralLayout provides standard padding for the layout.
     Views using this template should be built from scratch as
     the template only provides standard margin/padding without content. -->
<template>
  <div>
    <b-row v-if="adminPanelAvailable && showBar" class="position-sticky no-gutters bg-white border-bottom" :style="{ top: offsetTop + 'px', zIndex: 5 }">
      <b-container>
        <b-row>
          <b-button
            class="border-0 text-dark btn-back"
            variant="link"
            @click="backAction"
          >
            <i class="mr-2 text-decoration-none ac-icon ac-left-arrow ac-left-arrow-grey-dark"></i>
            <span class="btn-link text-dark">{{ backTitle }}</span>
          </b-button>
        </b-row>
      </b-container>
    </b-row>
    <alert-pane />
    <div v-if="adminPanelAvailable" class="pb-5">
      <slot name="default"></slot>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'

import AlertPane from '@/components/AlertPane.vue'
import { isAdmin } from '@/utils/permission'

export default {
  name: 'GeneralLayout',
  components: {
    AlertPane
  },

  props: {
    showBar: {
      type: Boolean,
      default: false
    },
    backTitle: {
      type: String,
      default: ''
    },
    backAction: {
      type: Function,
      default: () => {
        console.log('Please provide backAction')
      }
    }
  },

  data () {
    return {
      offsetTop: 0
    }
  },

  computed: {
    ...mapState('currentUser', [
      'user'
    ]),
    adminPanelAvailable () {
      return isAdmin(this.user)
    }
  },

  beforeCreate () {
    if (!this.user) {
      this.$store.dispatch('currentUser/get')
    }
  },

  mounted () {
    window.scrollTo(0, 0)
    // Set the top of elements to have sticky position
    // The offset top is according to the navbar height as sticky position requires
    // explictly set the top as the sticky threshold. In the case the backbar should
    // be sticky after the navbar height, which the threshold should be navbar height.
    this.offsetTop = this.$el.offsetTop
  }
}
</script>

<style scoped lang="scss">
button.btn-back {
    padding-top: 1rem;
    padding-bottom: 1rem;
}
</style>
