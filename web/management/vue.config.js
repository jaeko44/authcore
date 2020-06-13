const path = require('path')
const fs = require('fs')

const aliasPathSettings = {
  '@/router': path.resolve(__dirname, '/router/index.js')
}

module.exports = {
  devServer: {
    allowedHosts: [
      'authcore.dev',
      'server'
    ],
    public: 'authcore.dev:8001',
    sockPath: '/web/sockjs-node'
  },
  transpileDependencies: [
    /bootstrap-vue\/(?!node_modules)/
  ],
  outputDir: '../dist/web',
  publicPath: '/web/',
  assetsDir: 'assets',
  css: {
    loaderOptions: {
      sass: {
        data: '@import "@/css/_variable.scss";'
      }
    }
  },
  configureWebpack: {
    plugins: [
      {
        // TEMPORARY workaround for https://github.com/vuejs/vue-cli/issues/4400 and https://github.com/vuejs/vue-cli/issues/5372
        apply: compiler => {
          compiler.hooks.entryOption.tap('entry', () => {
            const clients = compiler.options.entry.app
            for (const index in clients) {
              if (clients[index].match(/sockjs-node/)) {
                clients[index] = clients[index].replace('authcore.dev:8001/sockjs-node', 'authcore.dev:8001&sockPath=/web/sockjs-node')
              }
            }
          })
        }
      }
    ]
  },
  chainWebpack: config => {
    config.resolve.alias.merge(aliasPathSettings)
    if (process.env.NODE_ENV === 'test') {
      // Configuration for generating coverage report using istanbul
      config.devtool('eval')
      config.module
        .rule('istanbul')
        .test(/\.(js|vue)$/)
        .include
        .end()
        .use('istanbul-instrumenter-loader')
        .loader('istanbul-instrumenter-loader')
        .options({ esModules: true })
        .end()
    }
    if (process.env.NODE_ENV === 'development') {
      const aliasconf = path.resolve(__dirname, '.package-alias.json')
      if (fs.existsSync(aliasconf)) {
        const jsoncontent = fs.readFileSync(aliasconf)
        const aliases = JSON.parse(jsoncontent)
        for (const [key, value] of Object.entries(aliases)) {
          const alias = path.resolve(__dirname, value)
          console.log('Adding webpack alias', key, alias)
          config.resolve.alias.set(key, alias)
        }
      }
    }
    return config
  }
}
