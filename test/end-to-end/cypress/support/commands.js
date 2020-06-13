// ***********************************************
// This example commands.js shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add("login", (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add("drag", { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add("dismiss", { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This is will overwrite an existing command --
// Cypress.Commands.overwrite("visit", (originalFn, url, options) => { ... })

// Add support for iframe command
// Reference from https://github.com/cypress-io/cypress/issues/136#issuecomment-417057086
Cypress.Commands.add('iframe', { prevSubject: 'element' }, ($iframe, src) => {
  Cypress.log({
    name: 'iframe',
    consoleProps() {
      return {
        iframe: $iframe,
      }
    },
  })
  let regexSrc = src !== undefined ? new RegExp(src) : undefined
  return new Cypress.Promise(resolve => {
    onIframeReady(
      $iframe,
      regexSrc,
      () => {
        console.log($iframe.attr('src'))
        // resolve($iframe)
        resolve($iframe.contents().find('body'))
      },
      () => {
        $iframe.on('load', () => {
          console.log('error case')
          resolve($iframe.contents().find('body'))
        })
      }
    )
  })
})

function onIframeReady($iframe, srcPath, successFn, errorFn) {
  try {
    const iCon = $iframe.first()[0].contentWindow,
          bl = 'about:blank',
          compl = 'complete'
    const callCallback = () => {
      try {
        const $con = $iframe.contents()
        if ($con.length === 0) {
          // https://git.io/vV8yU
          throw new Error('iframe inaccessible')
        }
        successFn($con)
      } catch (e) {
        // accessing contents failed
        errorFn()
      }
    }
    const observeOnload = () => {
      $iframe.on('load.jqueryMark', () => {
        try {
          const src = $iframe.attr('src').trim(),
                href = iCon.location.href
          if (href !== bl || src === bl || src === '') {
            $iframe.off('load.jqueryMark')
            callCallback()
          }
        } catch (e) {
          errorFn()
        }
      })
    }
    if (iCon.document.readyState === compl) {
      const src = $iframe.attr('src').trim(),
            href = iCon.location.href
      const matchSrcPath = srcPath !== undefined ? srcPath.test(src) : true
      console.log(matchSrcPath)
      if (href === bl && src !== bl && src !== '' && matchSrcPath) {
        observeOnload()
      } else {
        callCallback()
      }
    } else {
      observeOnload()
    }
  } catch (e) {
    // accessing contentWindow failed
    errorFn()
  }
}
