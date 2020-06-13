import _ from 'lodash'
import { assert } from 'chai'

import verifyAccountForm from '@/store/widgets/account/verify_account_form'

const {
  VERIFY_ACCOUNT_CONTACT_RESET,
  VERIFY_ACCOUNT_CONTACT_INIT,
  VERIFY_ACCOUNT_CONTACT_STARTED,
  VERIFY_ACCOUNT_CONTACT_COMPLETED,
  VERIFY_ACCOUNT_CONTACT_FAILED,
  CLEAR_STATES
} = verifyAccountForm.mutations
const { state: defaultState } = verifyAccountForm

suite('/store/widgets/account/verify_account_form.js', () => {
  suite('mutations', () => {
    let state
    beforeEach(() => {
      state = _.cloneDeep(defaultState)
    })
    test('should have correct default values', () => {
      assert.isFalse(state.loading)
      assert.isFalse(state.done)
      assert.isUndefined(state.error)
      assert.isUndefined(state.contactId)
    })
    suite('VERIFY_ACCOUNT_CONTACT_RESET', () => {
      test('should work properly', () => {
        // Preparing
        VERIFY_ACCOUNT_CONTACT_FAILED(state, new Error('imaginary error'))
        VERIFY_ACCOUNT_CONTACT_RESET(state)
        // Testing
        assert.isUndefined(state.error)
      })
    })
    suite('VERIFY_ACCOUNT_CONTACT_INIT', () => {
      suite('should work properly', () => {
        test('oldContact case', () => {
          // Preparing
          VERIFY_ACCOUNT_CONTACT_INIT(state, {
            contact: [{
              id: '1',
              type: 1,
              value: 'contact@example.com'
            }],
            oldContact: true
          })
          // Testing
          assert.equal(state.contact.id, '1')
          assert.equal(state.contact.value, 'contact@example.com')
          assert.isTrue(state.oldContactFlag)
        })
        test('Not oldContact case', () => {
          // Preparing
          VERIFY_ACCOUNT_CONTACT_INIT(state, {
            contact: [{
              id: '1',
              type: 1,
              value: 'contact@example.com'
            }],
            oldContact: false
          })
          // Testing
          assert.equal(state.contact.id, '1')
          assert.equal(state.contact.value, 'contact@example.com')
          assert.isFalse(state.oldContactFlag)
        })
      })
      test('More than one contact case', () => {
        // Preparing
        VERIFY_ACCOUNT_CONTACT_INIT(state, {
          contact: [{
            id: '1',
            type: 1,
            value: 'contact@example.com'
          }, {
            id: '2',
            type: 1,
            value: 'contact_new@example.com'
          }],
          oldContact: false
        })
        // Testing
        assert.isObject(state.contact)
        assert.isUndefined(state.contact.value)
        assert.isUndefined(state.contact.type)
        assert.isUndefined(state.oldContactFlag)
      })
    })
    suite('VERIFY_ACCOUNT_CONTACT_STARTED', () => {
      test('should work properly', () => {
        // Preparing
        VERIFY_ACCOUNT_CONTACT_STARTED(state)
        // Testing
        assert.isTrue(state.loading)
        assert.isFalse(state.done)
        assert.isUndefined(state.error)
      })
    })
    suite('VERIFY_ACCOUNT_CONTACT_COMPLETED', () => {
      test('should work properly', () => {
        // Preparing
        VERIFY_ACCOUNT_CONTACT_COMPLETED(state)
        // Testing
        assert.isFalse(state.loading)
        assert.isTrue(state.done)
        assert.isUndefined(state.error)
      })
    })
    suite('VERIFY_ACCOUNT_CONTACT_FAILED', () => {
      test('should work properly', () => {
        // Preparing
        VERIFY_ACCOUNT_CONTACT_FAILED(state, new Error('imaginary error'))
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
      })
      test('invalid verification code', () => {
        // Preparing
        VERIFY_ACCOUNT_CONTACT_FAILED(state, {
          response: {
            status: 401,
            statusText: 'Internal Server Error',
            obj: {
              code: 3,
              error: 'invalid argument',
              message: 'invalid argument'
            }
          }
        })
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
        assert.equal(state.error, 'verification.input.error.invalid_verification_code')
      })
      test('too frequent', () => {
        // Preparing
        VERIFY_ACCOUNT_CONTACT_FAILED(state, {
          response: {
            status: 429,
            statusText: '',
            obj: {
              code: 8,
              error: 'too many requests',
              message: 'too many requests'
            }
          }
        })
        // Testing
        assert.isFalse(state.loading)
        assert.exists(state.error)
        assert.equal(state.error, 'verification.input.error.too_frequent')
      })
    })
    suite('CLEAR_STATES', () => {
      test('should return default states', () => {
        // Preparing
        CLEAR_STATES(state)
        // Testing
        assert.isFalse(state.loading)
        assert.isFalse(state.done)
        assert.isUndefined(state.error)
        assert.isUndefined(state.contactId)
      })
    })
  })
})
