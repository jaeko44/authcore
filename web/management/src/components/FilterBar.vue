<template>
  <b-row id="filters">
    <b-col id="search" :cols="searchCol">
      <div class="border-theme-general d-inline-flex w-100">
        <b-button-group class="w-25">
          <b-dropdown class="w-100" toggle-class="d-inline-flex text-truncate text-decoration-none py-3 font:0.875rem text-grey text-hover-grey" variant="link" no-caret>
            <template v-slot:button-content>
              <span class="text-left text-truncate flex-grow-1">{{ queryKeyItem.name }}</span>
              <div>
                <i class="ml-2 ac ac-down-arrow"></i>
              </div>
            </template>
            <b-dropdown-item v-for="(item, index) in queryKeys" :key="index" @click="changeQuery(item)" :class="{ 'background-grey': queryKeyItem === item }">
              <span class="font:0.875rem" :class="{ 'text-primary': queryKeyItem === item }">{{ item.name }}</span>
            </b-dropdown-item>
          </b-dropdown>
        </b-button-group>
        <div class="vertical-stroke"></div>
        <b-form class="d-inline-flex w-75" @submit.prevent="submitEvent">
          <b-form-input
            class="h-100 font:0.875rem py-0 border-0 border-hover-0 border-focus-0"
            type="text"
            :placeholder='queryPlaceholder'
            v-model="mutatedQueryValue"
          />
          <b-button type="submit" variant="link" class="text-grey text-hover-grey">
            <i class="fas fa-search"></i>
          </b-button>
        </b-form>
      </div>
    </b-col>
    <b-col v-if="showSortSection" id="sort" cols="3">
      <div class="border-theme-general d-inline-flex w-100">
        <b-button-group class="w-100">
          <b-button-group class="w-75">
            <b-dropdown class="w-100" toggle-class="d-inline-flex text-truncate text-decoration-none py-3 font:0.875rem text-grey text-hover-grey" variant="link" no-caret>
              <template v-slot:button-content>
                <span class="text-left text-truncate flex-grow-1">{{ sortKeyItem.name }}</span>
                <div>
                  <i class="ml-2 ac ac-down-arrow"></i>
                </div>
              </template>
              <b-dropdown-item v-for="(item, index) in sortKeys" :key="index" @click="changeSort(item)" :class="{ 'background-grey': sortKeyItem === item }">
                <span class="font:0.875rem" :class="{ 'text-primary': sortKeyItem === item }">{{ item.name }}</span>
              </b-dropdown-item>
            </b-dropdown>
          </b-button-group>
          <div class="vertical-stroke"></div>
          <b-button class="w-25 p-0" variant="link" @click="changeOrder">
            <img alt="" src="@/assets/sort-icon.png" :class="{ 'rotateX-180': mutatedAscending }" />
          </b-button>
        </b-button-group>
      </div>
    </b-col>
  </b-row>
</template>

<script>
export default {
  // FilterBar based on row and column layout from grid
  name: 'FilterBar',
  components: {},
  props: {
    queryObject: {
      type: Object,
      default: () => {}
    },
    queryPlaceholder: {
      type: String,
      default: 'Placeholder'
    },
    queryKeys: {
      type: Array,
      required: true,
      // Check all items consist name and key fields of an object
      validator: (items) => {
        return items.every((item) =>
          !!item.name && !!item.key
        )
      }
    },
    // Provide sortKeys to show the sort section, otherwise it is hidden
    sortKeys: {
      type: Array,
      // Check all items consist name and key fields of an object
      // sortKeys can be empty as it is optional to show
      validator: (items) => {
        if (items.length > 0) {
          return items.every((item) =>
            !!item.name && !!item.key
          )
        }
        return true
      }
    }
  },

  data () {
    return {
      mutatedQueryKey: {},
      mutatedQueryValue: '',
      mutatedSortKey: {},
      mutatedAscending: false
    }
  },

  computed: {
    queryKeyItem () {
      return this.queryKeys.find(item => item.key === this.mutatedQueryKey.key)
    },
    sortKeyItem () {
      return this.sortKeys.find(item => item.key === this.mutatedSortKey.key)
    },
    mutatedQueryObject () {
      return {
        queryKey: this.mutatedQueryKey.key,
        queryValue: this.mutatedQueryValue,
        sortKey: this.mutatedSortKey.key,
        ascending: this.mutatedAscending
      }
    },
    showSortSection () {
      return !!this.sortKeys
    },
    searchCol () {
      return this.showSortSection ? 9 : 12
    }
  },

  watch: {
    // Handle the case when queryObject is updated from parent view
    // Case for query string updated by browser back/forward action.
    queryObject (newVal) {
      this.updateMutatedFields(newVal)
    }
  },

  created () {
    this.updateMutatedFields(this.queryObject)
  },

  methods: {
    submitEvent () {
      this.$emit('submit', this.mutatedQueryObject)
    },

    updateMutatedFields (obj) {
      this.mutatedQueryKey = this.queryKeys.find(item => item.key === obj.queryKey)
      this.mutatedQueryValue = obj.queryValue
      this.mutatedSortKey = this.sortKeys.find(item => item.key === obj.sortKey)
      this.mutatedAscending = obj.ascending
    },

    changeQuery (data) {
      this.mutatedQueryKey = data
    },

    changeSort (data) {
      this.mutatedSortKey = data
      this.$emit('submit', this.mutatedQueryObject)
    },

    changeOrder (data) {
      this.mutatedAscending = !this.mutatedAscending
      this.$emit('submit', this.mutatedQueryObject)
    }
  }
}
</script>

<style scoped lang="scss">
$vertical-stroke-margin: 0.6rem;

.vertical-stroke {
    margin-top: $vertical-stroke-margin;
    margin-bottom: $vertical-stroke-margin;
    margin-left: 0.2rem;
    border-right: 1px solid $bsq-grey-medium;
}
</style>

<style scoped lang="scss">
.fa-search {
    color: $bsq-grey-medium;
}
</style>
