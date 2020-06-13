// Automation for testing the reset password flow
// This should not be run in e2e test case as it intends to quicken first part of the flow only.
describe('Reset password from Sign In page', function () {
  it('Reset password flow', function () {
    cy.visit(Cypress.env('AUTHCORE_WEB_HOST'))

    cy.get('iframe')
      .iframe()
      .find('[data-cy=handle] > div.contact-input-include-label > input')
    // Handle for reset password, better change to be the one in local env
      .type(Cypress.env('test_account_username_for_login'))

    cy.get('iframe')
      .iframe()
      .find('[data-cy=reset_password]')
      .click()

    // Click send button to have reset password request
    cy.get('iframe')
      .iframe()
      .find('[data-cy=send]')
      .click()

    // Click OK button to confirm reset password request
    cy.get('iframe')
      .iframe()
      .find('[data-cy=ok]')
      .click()
  })
})
