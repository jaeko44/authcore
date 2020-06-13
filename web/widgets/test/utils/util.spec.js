import { camelCaseObject } from '@/utils/util'
import { assert } from 'chai'

suite('/utils/util.js', () => {
  suite('camelCaseObject', () => {
    test('Normal case', () => {
      const jsonObj = JSON.parse('{"under_score_case": "string"}')
      const result = camelCaseObject(jsonObj)
      assert.equal(Object.keys(result)[0], 'underScoreCase')
      assert.equal(result.underScoreCase, 'string')
    })
    test('No change case', () => {
      const jsonObj = JSON.parse('{"underScoreCase": 1}')
      const result = camelCaseObject(jsonObj)
      assert.equal(Object.keys(result)[0], 'underScoreCase')
      assert.equal(result.underScoreCase, 1)
    })
    test('Multiple under score case', () => {
      const jsonObj = JSON.parse('{"__under__score___case_": "string"}')
      const result = camelCaseObject(jsonObj)
      assert.equal(Object.keys(result)[0], 'underScoreCase')
      assert.equal(result.underScoreCase, 'string')
    })
  })
})
