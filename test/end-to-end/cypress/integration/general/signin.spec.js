describe('Login page', function () {
  it('Login flow', function () {
    cy.visit(Cypress.env('AUTHCORE_WEB_HOST'))

    cy.get('a.router-link-active')
      .should('contain', 'Sign In')

    cy.get('iframe')
      .iframe()
      .find('[data-cy=handle] > div.contact-input-include-label > input')
      .type(Cypress.env('test_account_username_for_login'))

    cy.get('iframe')
      .iframe()
      .find('button[type=submit]')
      .click()

    cy.get('iframe')
      .iframe()
      .find('[data-cy=password] > input')
      .type(Cypress.env('test_account_password'))

    cy.get('iframe')
      .iframe()
      .find('button[type=submit]')
      .click()

    cy.get('a.router-link-active', {
      // Apply longer timeout for CI environment
      timeout: 10000
    })
      .should('contain', 'Settings')
  })

  it('Register flow through Login widget', function () {
    cy.visit(Cypress.env('AUTHCORE_WEB_HOST'))

    cy.get('iframe')
      .iframe()
      .find('[data-cy=register-link]')
      .click()

    cy.wait(1000)

    const timeNow = Date.now()

    cy.get('iframe')
      .iframe()
      .find('[data-cy=contact] > div.contact-input-include-label > input')
      .type(Cypress.env('test_account_contact').replace('{TIMESTAMP}', timeNow))

    cy.get('iframe')
      .iframe()
      .find('[data-cy=password] > input')
      .type(Cypress.env('test_account_password'))

    cy.get('iframe')
      .iframe()
      .find('[data-cy=register]')
      .click()

    // TODO: Do not test the verification flow as there is no way to check the iframe src at the moment
    // Check https://github.com/cypress-io/cypress/issues/136 to see iframe related update

    // Wait time for API, workaround as verification flow cannot be tested
    cy.wait(5000)
  })
})
