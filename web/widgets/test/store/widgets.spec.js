import _ from 'lodash'
import { assert } from 'chai'

import widgets from '@/store/widgets'

const {
  SET_DISPLAY_MODE_STATE
} = widgets.mutations
const { state: defaultState } = widgets

suite('/store/widgets.js', () => {
  suite('mutations', () => {
    let state
    beforeEach(() => {
      state = _.cloneDeep(defaultState)
    })
    test('should have correct default values', () => {
      assert.isUndefined(state.displayMode)
    })
    suite('SET_DISPLAY_MODE_STATE', () => {
      test('should work properly', () => {
        // Preparing
        SET_DISPLAY_MODE_STATE(state, 'full-screen')
        // Testing
        assert.equal(state.displayMode, 'full-screen')
      })
    })
  })
})
