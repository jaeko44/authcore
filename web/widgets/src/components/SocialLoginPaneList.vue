<template>
  <b-row id="social-row" no-gutters class="mt-2 position-relative">
    <ul id="social-login-list" class="my-0 w-100">
      <li
        v-for="(item, index) in socialLoginList"
        :key="index"
        class="my-2"
      >
        <div v-if="index < 2 || expandedExtraOptions">
          <b-button
            block
            variant="normal"
            class="position-relative d-flex align-items-center px-2"
            @click="openOAuthWindow(item)"
          >
            <span
              class="social-logo-list-style"
              :class="['social-logo-list-style-' + item]"
            >
              <i class="icon" :class="['ac-' + item]"></i>
            </span>
            <i18n
              v-if="register"
              path="social_login_pane_list.text.register"
              tag="div"
              class="social-logo-list-wording-margin font-weight-normal"
            >
              <template #service>
                <span>{{ $t(`social_login_pane_list.text.${item}`) }}</span>
              </template>
            </i18n>
            <i18n
              v-else
              path="social_login_pane_list.text.sign_in"
              tag="div"
              class="social-logo-list-wording-margin font-weight-normal"
            >
              <template #service>
                <span>{{ $t(`social_login_pane_list.text.${item}`) }}</span>
              </template>
            </i18n>
          </b-button>
        </div>
      </li>
      <li v-if="!expandedExtraOptions" id="social-more" class="my-2">
        <b-button
          block
          ref="social-more-button"
          variant="normal"
          class="position-relative"
          @click="expandOptions"
        >
          <span>
            <i class="icon ac-ellipses"></i>
          </span>
        </b-button>
      </li>
    </ul>
  </b-row>
</template>

<script>
export default {
  name: 'SocialLoginPaneList',
  props: {
    // Decide to show what wordings in list button
    register: {
      type: Boolean,
      default: true
    },
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
      expandedExtraOptions: false
    }
  },

  methods: {
    expandOptions () {
      this.expandedExtraOptions = true
      // Unfocus when more button is clicked
      this.$refs['social-more-button'].blur()
    }
  }
}
</script>

<style scoped lang="scss">
.icon {
    display: inline;
    font-size: 1.4rem;
    /* Align the icon into center using font. Including Facebook, Twitter and Apple */
    line-height: 2.8rem;
}
.icon::before {
    display: block;
    width: 27px;
    height: 46px;
    margin: auto;
    background-position: center;
    background-repeat: no-repeat;
    background-size: contain;
}

.ac-google::before {
    content: ' ';
    background-image: url(data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiA/PjwhRE9DVFlQRSBzdmcgIFBVQkxJQyAnLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4nICAnaHR0cDovL3d3dy53My5vcmcvR3JhcGhpY3MvU1ZHLzEuMS9EVEQvc3ZnMTEuZHRkJz48c3ZnIGVuYWJsZS1iYWNrZ3JvdW5kPSJuZXcgMCAwIDQwMCA0MDAiIGhlaWdodD0iNDAwcHgiIGlkPSJMYXllcl8xIiB2ZXJzaW9uPSIxLjEiIHZpZXdCb3g9IjAgMCA0MDAgNDAwIiB3aWR0aD0iNDAwcHgiIHhtbDpzcGFjZT0icHJlc2VydmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6eGxpbms9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkveGxpbmsiPjxnPjxwYXRoIGQ9Ik0xNDIuOSwyNC4yQzk3LjYsMzkuNyw1OSw3My42LDM3LjUsMTE2LjVjLTcuNSwxNC44LTEyLjksMzAuNS0xNi4yLDQ2LjhjLTguMiw0MC40LTIuNSw4My41LDE2LjEsMTIwLjMgICBjMTIuMSwyNCwyOS41LDQ1LjQsNTAuNSw2Mi4xYzE5LjksMTUuOCw0MywyNy42LDY3LjYsMzQuMWMzMSw4LjMsNjQsOC4xLDk1LjIsMWMyOC4yLTYuNSw1NC45LTIwLDc2LjItMzkuNiAgIGMyMi41LTIwLjcsMzguNi00Ny45LDQ3LjEtNzcuMmM5LjMtMzEuOSwxMC41LTY2LDQuNy05OC44Yy01OC4zLDAtMTE2LjcsMC0xNzUsMGMwLDI0LjIsMCw0OC40LDAsNzIuNmMzMy44LDAsNjcuNiwwLDEwMS40LDAgICBjLTMuOSwyMy4yLTE3LjcsNDQuNC0zNy4yLDU3LjVjLTEyLjMsOC4zLTI2LjQsMTMuNi00MSwxNi4yYy0xNC42LDIuNS0yOS44LDIuOC00NC40LTAuMWMtMTQuOS0zLTI5LTkuMi00MS40LTE3LjkgICBjLTE5LjgtMTMuOS0zNC45LTM0LjItNDIuNi01Ny4xYy03LjktMjMuMy04LTQ5LjIsMC03Mi40YzUuNi0xNi40LDE0LjgtMzEuNSwyNy00My45YzE1LTE1LjQsMzQuNS0yNi40LDU1LjYtMzAuOSAgIGMxOC0zLjgsMzctMy4xLDU0LjYsMi4yYzE1LDQuNSwyOC44LDEyLjgsNDAuMSwyMy42YzExLjQtMTEuNCwyMi44LTIyLjgsMzQuMi0zNC4yYzYtNi4xLDEyLjMtMTIsMTguMS0xOC4zICAgYy0xNy4zLTE2LTM3LjctMjguOS01OS45LTM3LjFDMjI4LjIsMTAuNiwxODMuMiwxMC4zLDE0Mi45LDI0LjJ6IiBmaWxsPSIjRkZGRkZGIi8+PGc+PHBhdGggZD0iTTE0Mi45LDI0LjJjNDAuMi0xMy45LDg1LjMtMTMuNiwxMjUuMywxLjFjMjIuMiw4LjIsNDIuNSwyMSw1OS45LDM3LjFjLTUuOCw2LjMtMTIuMSwxMi4yLTE4LjEsMTguMyAgICBjLTExLjQsMTEuNC0yMi44LDIyLjgtMzQuMiwzNC4yYy0xMS4zLTEwLjgtMjUuMS0xOS00MC4xLTIzLjZjLTE3LjYtNS4zLTM2LjYtNi4xLTU0LjYtMi4yYy0yMSw0LjUtNDAuNSwxNS41LTU1LjYsMzAuOSAgICBjLTEyLjIsMTIuMy0yMS40LDI3LjUtMjcsNDMuOWMtMjAuMy0xNS44LTQwLjYtMzEuNS02MS00Ny4zQzU5LDczLjYsOTcuNiwzOS43LDE0Mi45LDI0LjJ6IiBmaWxsPSIjRUE0MzM1Ii8+PC9nPjxnPjxwYXRoIGQ9Ik0yMS40LDE2My4yYzMuMy0xNi4yLDguNy0zMiwxNi4yLTQ2LjhjMjAuMywxNS44LDQwLjYsMzEuNSw2MSw0Ny4zYy04LDIzLjMtOCw0OS4yLDAsNzIuNCAgICBjLTIwLjMsMTUuOC00MC42LDMxLjYtNjAuOSw0Ny4zQzE4LjksMjQ2LjcsMTMuMiwyMDMuNiwyMS40LDE2My4yeiIgZmlsbD0iI0ZCQkMwNSIvPjwvZz48Zz48cGF0aCBkPSJNMjAzLjcsMTY1LjFjNTguMywwLDExNi43LDAsMTc1LDBjNS44LDMyLjcsNC41LDY2LjgtNC43LDk4LjhjLTguNSwyOS4zLTI0LjYsNTYuNS00Ny4xLDc3LjIgICAgYy0xOS43LTE1LjMtMzkuNC0zMC42LTU5LjEtNDUuOWMxOS41LTEzLjEsMzMuMy0zNC4zLDM3LjItNTcuNWMtMzMuOCwwLTY3LjYsMC0xMDEuNCwwQzIwMy43LDIxMy41LDIwMy43LDE4OS4zLDIwMy43LDE2NS4xeiIgZmlsbD0iIzQyODVGNCIvPjwvZz48Zz48cGF0aCBkPSJNMzcuNSwyODMuNWMyMC4zLTE1LjcsNDAuNi0zMS41LDYwLjktNDcuM2M3LjgsMjIuOSwyMi44LDQzLjIsNDIuNiw1Ny4xYzEyLjQsOC43LDI2LjYsMTQuOSw0MS40LDE3LjkgICAgYzE0LjYsMywyOS43LDIuNiw0NC40LDAuMWMxNC42LTIuNiwyOC43LTcuOSw0MS0xNi4yYzE5LjcsMTUuMywzOS40LDMwLjYsNTkuMSw0NS45Yy0yMS4zLDE5LjctNDgsMzMuMS03Ni4yLDM5LjYgICAgYy0zMS4yLDcuMS02NC4yLDcuMy05NS4yLTFjLTI0LjYtNi41LTQ3LjctMTguMi02Ny42LTM0LjFDNjcsMzI4LjksNDkuNiwzMDcuNSwzNy41LDI4My41eiIgZmlsbD0iIzM0QTg1MyIvPjwvZz48L2c+PC9zdmc+);
}
.ac-facebook::before {
}
.ac-matters::before {
    width: 32px;
    content: ' ';
    background-image: url(data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB2aWV3Qm94PSIwIDAgMzE4LjYgMzE4LjYiPjxkZWZzPjxsaW5lYXJHcmFkaWVudCBpZD0iYSIgeDE9IjE1MC4zMSIgeTE9IjIxNC4wMSIgeDI9IjI5NC44NyIgeTI9IjEwMS4yNyIgZ3JhZGllbnRUcmFuc2Zvcm09Im1hdHJpeCgxLCAwLCAwLCAtMSwgMCwgMzIwKSIgZ3JhZGllbnRVbml0cz0idXNlclNwYWNlT25Vc2UiPjxzdG9wIG9mZnNldD0iMCIgc3RvcC1jb2xvcj0iI2Q3ZWFlMSIvPjxzdG9wIG9mZnNldD0iMSIgc3RvcC1jb2xvcj0iIzc5YjFhNiIvPjwvbGluZWFyR3JhZGllbnQ+PGxpbmVhckdyYWRpZW50IGlkPSJiIiB4MT0iNTMuNiIgeTE9IjIxOS44OCIgeDI9IjE3OC42MSIgeTI9IjkwLjE3IiBncmFkaWVudFRyYW5zZm9ybT0ibWF0cml4KDEsIDAsIDAsIC0xLCAwLCAzMjApIiBncmFkaWVudFVuaXRzPSJ1c2VyU3BhY2VPblVzZSI+PHN0b3Agb2Zmc2V0PSIwIiBzdG9wLWNvbG9yPSIjZjRlMmJjIi8+PHN0b3Agb2Zmc2V0PSIxIiBzdG9wLWNvbG9yPSIjYmY5ZjVlIi8+PC9saW5lYXJHcmFkaWVudD48L2RlZnM+PHRpdGxlPm1hdHRlcnM8L3RpdGxlPjxwYXRoIGQ9Ik0yMTksMjU2LjgyYTk3LjI5LDk3LjI5LDAsMSwwLTk3LjMtOTcuMjlBOTcuMjksOTcuMjksMCwwLDAsMjE5LDI1Ni44MloiIHN0eWxlPSJmaWxsLXJ1bGU6ZXZlbm9kZDtmaWxsOnVybCgjYSkiLz48cGF0aCBkPSJNMTEwLjg1LDI2Ny42M0ExMDguMSwxMDguMSwwLDEsMCwyLjc1LDE1OS41MywxMDguMSwxMDguMSwwLDAsMCwxMTAuODUsMjY3LjYzWiIgc3R5bGU9ImZpbGwtcnVsZTpldmVub2RkO2ZpbGw6dXJsKCNiKSIvPjwvc3ZnPg==);
}
.ac-twitter {
    font-size: 2rem;
}
.ac-apple {
    font-size: 2rem;
}

/* Settings for ellipses button in the page */
.ac-ellipses {
    &::before {
        line-height: 1;
        height: 24px;
    }
}

.scroll-row {
    display: flex;
    white-space: nowrap;
    overflow-y: none;
    overflow-x: auto;
}

.scroll {
    &.scroll-backward {
    }
    &.scroll-forward {
        a {
            right: 0;
        }
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
