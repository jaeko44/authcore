<template>
  <b-list-group-item
    class="generic-list-item"
  >
    <b-row align-v="center">
      <b-col cols="auto" class="pr-0 align-self-start">
        <div class="social-logo-rounded" :class="['social-logo-' + service]">
          <i class="icon" :class="['ac-' + service]"></i>
        </div>
      </b-col>
      <b-col>
        <b-row no-gutters>
          <b-col class="mb-2">
            <div>
              <h5 class="my-0">{{ serviceName }}</h5>
              <div>
                <i18n
                  v-if="factor"
                  path="manage_social_logins.list_item.text.last_used"
                  tag="span"
                >
                  <template #date>
                    <span>{{ factor.last_used_at | formatDatetime }}</span>
                  </template>
                </i18n>
                <span v-else>{{ $t('manage_social_logins.list_item.text.connect_now') }}</span>
              </div>
            </div>
          </b-col>
          <b-col cols="12" sm="auto">
            <b-button
              v-if="!factor"
              variant="primary"
              class="hover-remove"
              @click="openOAuthWindow(service)"
            >
              {{ $t('manage_social_logins.button.connect') }}
            </b-button>
            <b-button
              v-else
              :to="{ name: 'SocialLoginDelete', params: { id: factor.service } }"
              class="hover-remove"
              variant="danger"
            >
              {{ $t('manage_social_logins.button.disconnect') }}
            </b-button>
          </b-col>
        </b-row>
      </b-col>
    </b-row>
  </b-list-group-item>
</template>

<script>
import { mapState, mapActions } from 'vuex'

import { openOAuthWindow } from '@/utils/util'

export default {
  name: 'SocialLoginListItem',
  components: {},
  props: {
    factor: {
      type: Object,
      default: undefined
    },
    service: {
      type: String,
      default: ''
    }
  },

  data () {
    return {
      closeOAuthWindowFunc: null
    }
  },

  computed: {
    ...mapState('preferences', [
      'containerId'
    ]),
    ...mapState('authn', [
      'authnState',
      'error'
    ]),
    serviceName () {
      return this.$t(`manage_social_logins.list_item.title.${this.service}`)
    }
  },

  methods: {
    ...mapActions('authn', [
      'startIDPBinding'
    ]),
    async openOAuthWindow (idp) {
      this.closeOAuthWindowFunc = await openOAuthWindow(this.containerId, idp, async () => {
        const redirectURI = window.location.toString()
        await this.startIDPBinding({ idp, redirectURI })
        if (this.error) {
          throw new Error('error starting IDP binding')
        }
        if (this.authnState.status === 'IDP_BINDING') {
          const endpointUri = this.authnState.idp_authorization_url
          const state = this.authnState.state_token
          // Set the state token to allow resumption after redirection
          sessionStorage.setItem('io.authcore.authn_state.resume', state)
          return endpointUri
        }
        throw new Error('illegal state while starting IDP binding')
      })
    }
  }
}
</script>

<style scoped lang="scss">
.icon {
    display: inline;
    font-size: 1.25rem;
    /* Align the icon into center */
    line-height: 2.5rem;
}
.icon::before {
    display: block;
    width: 18px;
    height: 38px;
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
    width: 24px;
    content: ' ';
    background-image: url(data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB2aWV3Qm94PSIwIDAgMzE4LjYgMzE4LjYiPjxkZWZzPjxsaW5lYXJHcmFkaWVudCBpZD0iYSIgeDE9IjE1MC4zMSIgeTE9IjIxNC4wMSIgeDI9IjI5NC44NyIgeTI9IjEwMS4yNyIgZ3JhZGllbnRUcmFuc2Zvcm09Im1hdHJpeCgxLCAwLCAwLCAtMSwgMCwgMzIwKSIgZ3JhZGllbnRVbml0cz0idXNlclNwYWNlT25Vc2UiPjxzdG9wIG9mZnNldD0iMCIgc3RvcC1jb2xvcj0iI2Q3ZWFlMSIvPjxzdG9wIG9mZnNldD0iMSIgc3RvcC1jb2xvcj0iIzc5YjFhNiIvPjwvbGluZWFyR3JhZGllbnQ+PGxpbmVhckdyYWRpZW50IGlkPSJiIiB4MT0iNTMuNiIgeTE9IjIxOS44OCIgeDI9IjE3OC42MSIgeTI9IjkwLjE3IiBncmFkaWVudFRyYW5zZm9ybT0ibWF0cml4KDEsIDAsIDAsIC0xLCAwLCAzMjApIiBncmFkaWVudFVuaXRzPSJ1c2VyU3BhY2VPblVzZSI+PHN0b3Agb2Zmc2V0PSIwIiBzdG9wLWNvbG9yPSIjZjRlMmJjIi8+PHN0b3Agb2Zmc2V0PSIxIiBzdG9wLWNvbG9yPSIjYmY5ZjVlIi8+PC9saW5lYXJHcmFkaWVudD48L2RlZnM+PHRpdGxlPm1hdHRlcnM8L3RpdGxlPjxwYXRoIGQ9Ik0yMTksMjU2LjgyYTk3LjI5LDk3LjI5LDAsMSwwLTk3LjMtOTcuMjlBOTcuMjksOTcuMjksMCwwLDAsMjE5LDI1Ni44MloiIHN0eWxlPSJmaWxsLXJ1bGU6ZXZlbm9kZDtmaWxsOnVybCgjYSkiLz48cGF0aCBkPSJNMTEwLjg1LDI2Ny42M0ExMDguMSwxMDguMSwwLDEsMCwyLjc1LDE1OS41MywxMDguMSwxMDguMSwwLDAsMCwxMTAuODUsMjY3LjYzWiIgc3R5bGU9ImZpbGwtcnVsZTpldmVub2RkO2ZpbGw6dXJsKCNiKSIvPjwvc3ZnPg==);
}
.ac-apple {
    color: white;
}
</style>
