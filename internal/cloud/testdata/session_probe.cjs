// Execute the production probe against a captured/mocked 115 response.
const { readFileSync } = require('node:fs')
const vm = require('node:vm')
const input = JSON.parse(readFileSync(0, 'utf8'))
const context = {
  location: { href: 'https://115.com/?mode=wangpan' },
  USER_ID: input.userId,
  window: { browserInterface: { ChromeGetLixianKey() {}, CreateDownloadTask() {} } },
  AbortSignal,
  fetch: async (url, options) => {
    if (url !== 'https://my.115.com/?ct=ajax&ac=nav' || options.credentials !== 'include' || options.cache !== 'no-store' || !options.signal) {
      throw new Error('invalid session probe request')
    }
    if (input.failure) throw new Error(input.failure)
    return {
      ok: input.status === 200, status: input.status,
      json: async () => {
        if (input.html) throw new SyntaxError('Unexpected token <')
        return input.response
      }
    }
  }
}
context.browserInterface = context.window.browserInterface
vm.runInNewContext(input.expression, context).then(result => process.stdout.write(JSON.stringify(result))).catch(error => {
  console.error(error)
  process.exitCode = 1
})
