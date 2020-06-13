<template>
  <b-row id="social-row" no-gutters class="position-relative" :class="{ 'scroll-row': socialScrollRow, 'justify-content-center': !socialScrollRow }">
    <span class="position-absolute h-100 scroll scroll-backward flex-center-vertically absolute-center-vertically" :style="backwardArrowStyle">
      <a
        class="w-100 h-100"
        href="#"
        @click="backwardScroll"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" role="img" class="icon h-100">
          <path d="M21.5265 8.77171C22.1578 8.13764 22.1578 7.10962 21.5265 6.47555C20.8951 5.84148 19.8714 5.84148 19.24 6.47555L11.9999 13.7465L4.75996 6.47573C4.12858 5.84166 3.10492 5.84166 2.47354 6.47573C1.84215 7.10979 1.84215 8.13782 2.47354 8.77188L10.8332 17.1671C10.8408 17.1751 10.8486 17.183 10.8565 17.1909C11.0636 17.399 11.313 17.5388 11.577 17.6103C11.5834 17.6121 11.5899 17.6138 11.5964 17.6154C12.132 17.7536 12.7242 17.6122 13.1435 17.1911C13.1539 17.1807 13.1641 17.1702 13.1742 17.1596L21.5265 8.77171Z" fill="black" style="transform: rotate(90deg); transform-origin: center;"></path>
        </svg>
      </a>
    </span>
    <span class="position-absolute h-100 scroll scroll-forward flex-center-vertically absolute-center-vertically" :style="forwardArrowStyle">
      <a
        class="w-100 h-100"
        href="#"
        @click="forwardScroll"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" role="img" class="icon h-100">
          <path d="M21.5265 8.77171C22.1578 8.13764 22.1578 7.10962 21.5265 6.47555C20.8951 5.84148 19.8714 5.84148 19.24 6.47555L11.9999 13.7465L4.75996 6.47573C4.12858 5.84166 3.10492 5.84166 2.47354 6.47573C1.84215 7.10979 1.84215 8.13782 2.47354 8.77188L10.8332 17.1671C10.8408 17.1751 10.8486 17.183 10.8565 17.1909C11.0636 17.399 11.313 17.5388 11.577 17.6103C11.5834 17.6121 11.5899 17.6138 11.5964 17.6154C12.132 17.7536 12.7242 17.6122 13.1435 17.1911C13.1539 17.1807 13.1641 17.1702 13.1742 17.1596L21.5265 8.77171Z" fill="black" style="transform: rotate(270deg); transform-origin: center;"></path>
        </svg>
      </a>
    </span>
    <ul
      ref="scrollList"
      id="social-login-list"
      :class="{ 'scrolling' : socialScrollRow }"
      class="d-flex my-0"
      v-on:scroll.passive="socialLoginScroll"
    >
      <li
        v-for="(item, index) in socialLoginList"
        :key="index"
        class="px-0 flex-center-vertically"
        cols="auto"
        :style="[index === 0 ? { 'margin-right': marginForSocialLogo + 'px' } : { 'margin-left': marginForSocialLogo + 'px', 'margin-right': marginForSocialLogo + 'px' }]"
      >
        <b-link
          class="social-logo"
          :class="['social-logo-' + item]"
          href="#"
          @click="openOAuthWindow(item)"
        >
          <i class="icon ac-icon" :class="['ac-' + item]"></i>
        </b-link>
      </li>
    </ul>
  </b-row>
</template>

<script>
import { Tween, update } from 'es6-tween'

export default {
  name: 'SocialLoginPaneGrid',
  props: {
    socialLoginList: {
      type: Array,
      default: () => []
    },
    // Using parent component to serve as controller for open OAuth window function.
    openOAuthWindow: {
      type: Function,
      required: true
    }
  },

  data () {
    return {
      marginForSocialLogo: 16,
      socialScrollRow: false,
      scrollListBackwardEnabled: false,
      scrollListForwardEnabled: true,
      scrollLeft: 0,
      tweenedScrollLeft: { value: 0 }
    }
  },

  computed: {
    showBackwardScroll () {
      return this.socialScrollRow && this.scrollListBackwardEnabled
    },
    showForwardScroll () {
      return this.socialScrollRow && this.scrollListForwardEnabled
    },
    backwardArrowStyle () {
      if (this.showBackwardScroll) {
        return {
          zIndex: 1
        }
      }
      return {
        display: 'none'
      }
    },
    forwardArrowStyle () {
      if (this.showForwardScroll) {
        return {
          zIndex: 1
        }
      }
      return {
        display: 'none'
      }
    }
  },

  watch: {
    scrollLeft (newVal) {
      function animate () {
        if (update()) {
          requestAnimationFrame(animate)
        }
      }

      new Tween(this.tweenedScrollLeft)
        .to({ value: newVal }, 300)
        .on('update', ({ value }) => {
          this.$refs.scrollList.scrollLeft = value
        })
        .start()
      animate()
    }
  },

  mounted () {
    window.addEventListener('resize', this.resizeMarginForSocialLogo)
  },
  updated () {
    this.resizeMarginForSocialLogo()
  },
  beforeDestroy () {
    window.removeEventListener('resize', this.resizeMarginForSocialLogo)
  },

  methods: {
    resizeMarginForSocialLogo () {
      const socialLoginListElement = document.getElementById('social-login-list')
      const iconElement = socialLoginListElement.firstChild
      const rowElement = document.getElementById('social-row')
      if (iconElement !== null && rowElement !== null) {
        let socialRowWidthRequirement = (iconElement.clientWidth + this.marginForSocialLogo * 2) * this.socialLoginList.length
        if (socialRowWidthRequirement > rowElement.clientWidth) {
          this.socialScrollRow = true
        }
        // Build new margin for social logo
        if (this.socialScrollRow) {
          if (window.innerWidth < 370) {
            this.marginForSocialLogo = window.innerWidth / 28
          } else if (window.innerWidth < 410) {
            this.marginForSocialLogo = window.innerWidth / 32
          } else if (window.innerWidth < 480) {
            this.marginForSocialLogo = window.innerWidth / 25
          }
        }

        socialRowWidthRequirement = (iconElement.clientWidth + this.marginForSocialLogo * 2) * 5
        if (socialRowWidthRequirement <= rowElement.clientWidth) {
          this.socialScrollRow = false
          this.marginForSocialLogo = 16
        }
      }
    },
    socialLoginScroll (ev) {
      this.scrollListBackwardEnabled = ev.target.scrollLeft > 0
      this.scrollListForwardEnabled = ev.target.scrollWidth - (ev.target.clientWidth + ev.target.scrollLeft) > 0
    },
    backwardScroll () {
      this.scrollLeft = this.$refs.scrollList.scrollLeft - this.$refs.scrollList.clientWidth + 20
    },
    forwardScroll () {
      this.scrollLeft = this.$refs.scrollList.scrollLeft + this.$refs.scrollList.clientWidth - 20
    }
  }
}
</script>

<style scoped lang="scss">
.icon {
    display: inline;
    font-size: 1.4rem;
    /* Align the icon into center */
    line-height: 3rem;
}
.icon::before {
    display: block;
    width: 27px;
    height: 48px;
    margin: auto;
    background-position: center;
    background-repeat: no-repeat;
    background-size: contain;
}

.ac-google::before {
    content: ' ';
    background-image: url(data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiA/PjwhRE9DVFlQRSBzdmcgIFBVQkxJQyAnLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4nICAnaHR0cDovL3d3dy53My5vcmcvR3JhcGhpY3MvU1ZHLzEuMS9EVEQvc3ZnMTEuZHRkJz48c3ZnIGVuYWJsZS1iYWNrZ3JvdW5kPSJuZXcgMCAwIDQwMCA0MDAiIGhlaWdodD0iNDAwcHgiIGlkPSJMYXllcl8xIiB2ZXJzaW9uPSIxLjEiIHZpZXdCb3g9IjAgMCA0MDAgNDAwIiB3aWR0aD0iNDAwcHgiIHhtbDpzcGFjZT0icHJlc2VydmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiPjxnPjxwYXRoIGQ9Ik0xNDIuOSwyNC4yQzk3LjYsMzkuNyw1OSw3My42LDM3LjUsMTE2LjVjLTcuNSwxNC44LTEyLjksMzAuNS0xNi4yLDQ2LjhjLTguMiw0MC40LTIuNSw4My41LDE2LjEsMTIwLjMgICBjMTIuMSwyNCwyOS41LDQ1LjQsNTAuNSw2Mi4xYzE5LjksMTUuOCw0MywyNy42LDY3LjYsMzQuMWMzMSw4LjMsNjQsOC4xLDk1LjIsMWMyOC4yLTYuNSw1NC45LTIwLDc2LjItMzkuNiAgIGMyMi41LTIwLjcsMzguNi00Ny45LDQ3LjEtNzcuMmM5LjMtMzEuOSwxMC41LTY2LDQuNy05OC44Yy01OC4zLDAtMTE2LjcsMC0xNzUsMGMwLDI0LjIsMCw0OC40LDAsNzIuNmMzMy44LDAsNjcuNiwwLDEwMS40LDAgICBjLTMuOSwyMy4yLTE3LjcsNDQuNC0zNy4yLDU3LjVjLTEyLjMsOC4zLTI2LjQsMTMuNi00MSwxNi4yYy0xNC42LDIuNS0yOS44LDIuOC00NC40LTAuMWMtMTQuOS0zLTI5LTkuMi00MS40LTE3LjkgICBjLTE5LjgtMTMuOS0zNC45LTM0LjItNDIuNi01Ny4xYy03LjktMjMuMy04LTQ5LjIsMC03Mi40YzUuNi0xNi40LDE0LjgtMzEuNSwyNy00My45YzE1LTE1LjQsMzQuNS0yNi40LDU1LjYtMzAuOSAgIGMxOC0zLjgsMzctMy4xLDU0LjYsMi4yYzE1LDQuNSwyOC44LDEyLjgsNDAuMSwyMy42YzExLjQtMTEuNCwyMi44LTIyLjgsMzQuMi0zNC4yYzYtNi4xLDEyLjMtMTIsMTguMS0xOC4zICAgYy0xNy4zLTE2LTM3LjctMjguOS01OS45LTM3LjFDMjI4LjIsMTAuNiwxODMuMiwxMC4zLDE0Mi45LDI0LjJ6IiBmaWxsPSIjRkZGRkZGIi8+PGc+PHBhdGggZD0iTTE0Mi45LDI0LjJjNDAuMi0xMy45LDg1LjMtMTMuNiwxMjUuMywxLjFjMjIuMiw4LjIsNDIuNSwyMSw1OS45LDM3LjFjLTUuOCw2LjMtMTIuMSwxMi4yLTE4LjEsMTguMyAgICBjLTExLjQsMTEuNC0yMi44LDIyLjgtMzQuMiwzNC4yYy0xMS4zLTEwLjgtMjUuMS0xOS00MC4xLTIzLjZjLTE3LjYtNS4zLTM2LjYtNi4xLTU0LjYtMi4yYy0yMSw0LjUtNDAuNSwxNS41LTU1LjYsMzAuOSAgICBjLTEyLjIsMTIuMy0yMS40LDI3LjUtMjcsNDMuOWMtMjAuMy0xNS44LTQwLjYtMzEuNS02MS00Ny4zQzU5LDczLjYsOTcuNiwzOS43LDE0Mi45LDI0LjJ6IiBmaWxsPSIjRUE0MzM1Ii8+PC9nPjxnPjxwYXRoIGQ9Ik0yMS40LDE2My4yYzMuMy0xNi4yLDguNy0zMiwxNi4yLTQ2LjhjMjAuMywxNS44LDQwLjYsMzEuNSw2MSw0Ny4zYy04LDIzLjMtOCw0OS4yLDAsNzIuNCAgICBjLTIwLjMsMTUuOC00MC42LDMxLjYtNjAuOSw0Ny4zQzE4LjksMjQ2LjcsMTMuMiwyMDMuNiwyMS40LDE2My4yeiIgZmlsbD0iI0ZCQkMwNSIvPjwvZz48Zz48cGF0aCBkPSJNMjAzLjcsMTY1LjFjNTguMywwLDExNi43LDAsMTc1LDBjNS44LDMyLjcsNC41LDY2LjgtNC43LDk4LjhjLTguNSwyOS4zLTI0LjYsNTYuNS00Ny4xLDc3LjIgICAgYy0xOS43LTE1LjMtMzkuNC0zMC42LTU5LjEtNDUuOWMxOS41LTEzLjEsMzMuMy0zNC4zLDM3LjItNTcuNWMtMzMuOCwwLTY3LjYsMC0xMDEuNCwwQzIwMy43LDIxMy41LDIwMy43LDE4OS4zLDIwMy43LDE2NS4xeiIgZmlsbD0iIzQyODVGNCIvPjwvZz48Zz48cGF0aCBkPSJNMzcuNSwyODMuNWMyMC4zLTE1LjcsNDAuNi0zMS41LDYwLjktNDcuM2M3LjgsMjIuOSwyMi44LDQzLjIsNDIuNiw1Ny4xYzEyLjQsOC43LDI2LjYsMTQuOSw0MS40LDE3LjkgICAgYzE0LjYsMywyOS43LDIuNiw0NC40LDAuMWMxNC42LTIuNiwyOC43LTcuOSw0MS0xNi4yYzE5LjcsMTUuMywzOS40LDMwLjYsNTkuMSw0NS45Yy0yMS4zLDE5LjctNDgsMzMuMS03Ni4yLDM5LjYgICAgYy0zMS4yLDcuMS02NC4yLDcuMy05NS4yLTFjLTI0LjYtNi41LTQ3LjctMTguMi02Ny42LTM0LjFDNjcsMzI4LjksNDkuNiwzMDcuNSwzNy41LDI4My41eiIgZmlsbD0iIzM0QTg1MyIvPjwvZz48L2c+PC9zdmc+);
}
.ac-matters::before {
    width: 35px;
    content: ' ';
    background-image: url(data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB2aWV3Qm94PSIwIDAgMzE4LjYgMzE4LjYiPjxkZWZzPjxsaW5lYXJHcmFkaWVudCBpZD0iYSIgeDE9IjE1MC4zMSIgeTE9IjIxNC4wMSIgeDI9IjI5NC44NyIgeTI9IjEwMS4yNyIgZ3JhZGllbnRUcmFuc2Zvcm09Im1hdHJpeCgxLCAwLCAwLCAtMSwgMCwgMzIwKSIgZ3JhZGllbnRVbml0cz0idXNlclNwYWNlT25Vc2UiPjxzdG9wIG9mZnNldD0iMCIgc3RvcC1jb2xvcj0iI2Q3ZWFlMSIvPjxzdG9wIG9mZnNldD0iMSIgc3RvcC1jb2xvcj0iIzc5YjFhNiIvPjwvbGluZWFyR3JhZGllbnQ+PGxpbmVhckdyYWRpZW50IGlkPSJiIiB4MT0iNTMuNiIgeTE9IjIxOS44OCIgeDI9IjE3OC42MSIgeTI9IjkwLjE3IiBncmFkaWVudFRyYW5zZm9ybT0ibWF0cml4KDEsIDAsIDAsIC0xLCAwLCAzMjApIiBncmFkaWVudFVuaXRzPSJ1c2VyU3BhY2VPblVzZSI+PHN0b3Agb2Zmc2V0PSIwIiBzdG9wLWNvbG9yPSIjZjRlMmJjIi8+PHN0b3Agb2Zmc2V0PSIxIiBzdG9wLWNvbG9yPSIjYmY5ZjVlIi8+PC9saW5lYXJHcmFkaWVudD48L2RlZnM+PHRpdGxlPm1hdHRlcnM8L3RpdGxlPjxwYXRoIGQ9Ik0yMTksMjU2LjgyYTk3LjI5LDk3LjI5LDAsMSwwLTk3LjMtOTcuMjlBOTcuMjksOTcuMjksMCwwLDAsMjE5LDI1Ni44MloiIHN0eWxlPSJmaWxsLXJ1bGU6ZXZlbm9kZDtmaWxsOnVybCgjYSkiLz48cGF0aCBkPSJNMTEwLjg1LDI2Ny42M0ExMDguMSwxMDguMSwwLDEsMCwyLjc1LDE1OS41MywxMDguMSwxMDguMSwwLDAsMCwxMTAuODUsMjY3LjYzWiIgc3R5bGU9ImZpbGwtcnVsZTpldmVub2RkO2ZpbGw6dXJsKCNiKSIvPjwvc3ZnPg==);
}
.ac-twitter {
    font-size: 2rem;
}
.ac-apple {
    font-size: 2rem;
    color: white;
}

.scroll-row {
    display: flex;
    white-space: nowrap;
    overflow-y: none;
    overflow-x: auto;
}

.scroll {
    width: 24px;
    &.scroll-backward {
        text-align: left;
    }
    &.scroll-forward {
        text-align: right;
        right: 0;
    }
}

.scrolling {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    -ms-overflow-style: -ms-autohiding-scrollbar;

    &::-webkit-scrollbar {
        display: none;
    }
}
</style>
