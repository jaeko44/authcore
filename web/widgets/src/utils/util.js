import mapKeys from 'lodash-es/mapKeys'
import camelCase from 'lodash-es/camelCase'

// Detect if the browser is Facebook in-app browser.
// ref: https://stackoverflow.com/a/32348687
function isFacebook () {
  var ua = navigator.userAgent || navigator.vendor || window.opera
  return (ua.indexOf('FBAN') > -1) || (ua.indexOf('FBAV') > -1)
}

// Detect if the browser is in standalone mode.
// ref: https://web.dev/customize-install/
function isStandaloneMode () {
  return window.navigator.standalone || (window.matchMedia('(display-mode: standalone)').matches)
}

// redirectTo function checks whether the parent window can be redirected directly
// without using postMessage
export function redirectTo (urlString, containerId) {
  const urlObj = new URL(urlString)
  if (urlObj.origin === window.location.origin) {
    parent.document.location = urlString
  } else {
    parent.postMessage({
      type: 'AuthCore_redirectSuccessUrl',
      data: {
        containerId: containerId,
        redirectUrl: urlString
      }
    }, document.referrer)
  }
}

// inMobile function return whether the user agent of instance is in mobile or not
// for safely open a pop-up tab / window.
export function inMobile () {
  return /Mobi/.test(navigator.userAgent) || isFacebook() || isStandaloneMode()
}

// camelCaseObject function return convert object keys into camelCase
export function camelCaseObject (obj) {
  return mapKeys(obj, (value, key) => camelCase(key))
}
const oauthWindowSizes = {
  google: { height: 600, width: 400 },
  facebook: { height: 500, width: 400 },
  apple: { height: 750, width: 800 },
  matters: { height: 700, width: 500 },
  twitter: { height: 550, width: 720 }
}

// openOAuthWindow open a window for OAuth flow.
export async function openOAuthWindow (containerId, service, f) {
  let w
  if (!inMobile()) {
    w = window.open('about:blank', 'window', `height=${oauthWindowSizes[service].height},width=${oauthWindowSizes[service].width},menubar=no`)
  }
  const closer = function () {
    if (w) {
      w.close()
    }
  }
  try {
    const url = await f()
    if (w) {
      w.location = url
    } else {
      redirectTo(url, containerId)
    }
  } catch (e) {
    console.error('error when opening OAuth window', e)
    closer()
  }
  return closer
}
