<template>
  <div>
    <b-table
      show-empty
      fixed
      cellspacing="0"
      :class="{ 'item-clickable': $listeners['row-clicked'], 'theme-general': !themeFull, 'theme-full': themeFull, 'theme-general-with-pagniation': !themeFull && showPagination }"
      :loading="loading"
      :items="itemsDependsOnLoading"
      :fields="fields"
      :empty-text="emptyText"
      v-on:row-clicked="$emit('row-clicked', $event)"
    >
      <template v-slot:cell()="data">
        <with-loading-content
          :loading="loading"
          :title="data.value"
          class="text-break"
        >
          {{ data.value }}
        </with-loading-content>
      </template>
      <template v-slot:cell(actions)="data">
        <with-loading-content :loading="loading">
          <slot v-bind:data="data.item"></slot>
        </with-loading-content>
      </template>
    </b-table>
    <b-row v-if="showPagination" align-v="center" :class="{ 'px-3': !themeFull }">
      <b-col cols="6">
        <b-dropdown toggle-class="d-inline-flex text-truncate text-decoration-none py-2 font:0.875rem text-grey text-hover-grey border" variant="link" no-caret>
          <template v-slot:button-content>
            <span class="text-left text-truncate flex-grow-1">{{ totalItemsDescription }}</span>
            <div>
              <i class="ml-2 ac ac-down-arrow"></i>
            </div>
          </template>
          <b-dropdown-item-button @click="$emit('first-page-click')">
            {{ $t('data_table.text.first_page') }}
          </b-dropdown-item-button>
        </b-dropdown>
      </b-col>
      <b-col cols="6" class="text-right">
        <pagination
          class="list-with-table-padding-right"
          :previous-button-text="previousButtonText"
          :next-button-text="nextButtonText"
          :previous-page-token="previousPageToken"
          :next-page-token="nextPageToken"
          @previous-click="$emit('previous-click')"
          @next-click="$emit('next-click')"
        />
      </b-col>
    </b-row>
  </div>
</template>

<script>
import Pagination from '@/components/Pagination.vue'
import WithLoadingContent from '@/components/WithLoadingContent.vue'

export default {
  name: 'DataTable',
  components: {
    WithLoadingContent,
    Pagination
  },
  props: {
    themeFull: {
      type: Boolean,
      default: false
    },
    loading: {
      type: Boolean,
      default: false
    },
    items: {
      type: Array
    },
    fields: {
      type: Array
    },
    previousPageToken: {
      type: String,
      default: undefined
    },
    nextPageToken: {
      type: String,
      default: undefined
    },
    previousButtonText: {
      type: String,
      default: ''
    },
    nextButtonText: {
      type: String,
      default: ''
    },
    totalSize: {
      type: Number,
      default: -1
    },
    totalItemsDescription: {
      type: String,
      default: ''
    },
    itemNumberOnLoading: {
      type: Number,
      default: 3
    },
    emptyText: {
      type: String
    }
  },

  data () {
    return {
      defaultItems: new Array(this.itemNumberOnLoading).fill({})
    }
  },

  computed: {
    showPagination () {
      return this.previousPageToken !== undefined || this.nextPageToken !== undefined
    },
    // itemsDependsOnLoading is used to show loading layout when it is in loading status
    // defaultItems are empty items. The loading status trigger the item cell to have loading layout
    itemsDependsOnLoading () {
      return this.loading ? this.defaultItems : this.items
    }
  }
}
</script>

<style lang="scss">
@mixin table-general($border-width, $border-color) {
    tbody {
        tr {
            transition: background 0.3s cubic-bezier(0.25, 0.8, 0.5, 1);
            will-change: background;
        }
    }
    th {
        padding: $table-general-td-padding-y $table-general-td-padding-x;
        border-bottom: $border-width solid $border-color !important;
    }
    td {
        padding: $table-general-td-padding-y $table-general-td-padding-x;
        border-top: 0;
        border-bottom: $border-width solid $border-color;
        vertical-align: middle;
    }

    tr {
        &:focus {
            outline: 0;
        }
    }

    &.item-clickable {
        tbody {
            tr {
                &:hover {
                    background-color: $bsq-grey-light;
                }
            }
        }
    }
}

table.theme-general {
    @include table-general($general-border-size, $bsq-grey-medium);
    margin-bottom: 0;
    thead th {
        border-bottom: $border-width solid $bsq-grey-medium;
    }
    th {
        color: $bsq-text-grey;
        font-weight: normal;
    }

    tr:last-child td {
        border-bottom: 0;
        border-bottom-left-radius: 3px;
        border-bottom-right-radius: 3px;
    }

    &.theme-general-with-pagniation {
        margin-bottom: 1rem;
        tr:last-child td {
            border-bottom: $border-width solid $bsq-grey-medium;
        }
    }
}

// Apply cellspacing="0" in attribute
table.theme-full {
    $border-radius: 5px;
    @include table-general($general-border-size, $bsq-grey-medium);
    border-collapse: initial;
    td {
        vertical-align: middle;
    }
    th {
        border-top: $general-border-size solid $bsq-grey-medium;
        color: $bsq-text-grey;
        font-weight: normal;
    }
    tr th:first-child, td:first-child {
        border-left: $general-border-size solid $bsq-grey-medium;
    }
    tr th:last-child, td:last-child {
        border-right: $general-border-size solid $bsq-grey-medium;
    }

    tr:first-child th:first-child {
        border-top-left-radius: $border-radius;
    }
    tr:first-child th:last-child {
        border-top-right-radius: $border-radius;
    }
    tr:last-child td:first-child {
        border-bottom-left-radius: $border-radius;
    }
    tr:last-child td:last-child {
        border-bottom-right-radius: $border-radius;
    }
}

table.item-clickable {
    tbody {
        tr {
            &:hover {
                cursor: pointer;
            }
        }
    }
}

.list-with-table-padding {
    &-left {
        padding-left: ($general-border-size + $table-general-td-padding-x);
    }

    &-right {
        padding-right: ($general-border-size + $table-general-td-padding-x);
    }
}
</style>
