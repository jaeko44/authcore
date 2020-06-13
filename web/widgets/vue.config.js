const MomentLocalesPlugin = require('moment-locales-webpack-plugin')

const path = require('path')
const fs = require('fs')
module.exports = {
  devServer: {
    allowedHosts: [
      'authcore.dev'
    ],
    public: 'authcore.dev:8001',
    sockPath: '/widgets/sockjs-node'
  },
  transpileDependencies: [
    /bootstrap-vue\/(?!node_modules)/
  ],
  outputDir: '../dist/widgets',
  publicPath: '/widgets/',
  assetsDir: 'assets',
  configureWebpack: {
    plugins: [
      new MomentLocalesPlugin(),
      {
        // TEMPORARY workaround for https://github.com/vuejs/vue-cli/issues/4400 and https://github.com/vuejs/vue-cli/issues/5372
        apply: compiler => {
          compiler.hooks.entryOption.tap('entry', () => {
            const clients = compiler.options.entry.app
            for (const index in clients) {
              if (clients[index].match(/sockjs-node/)) {
                clients[index] = clients[index].replace('authcore.dev:8001/sockjs-node', 'authcore.dev:8001&sockPath=/widgets/sockjs-node')
              }
            }
          })
        }
      }
    ],
    optimization: {
      splitChunks: {
        cacheGroups: {
          vendor: {
            // Separate zxcvbn package out and name as password-vendor
            name: 'password-vendor',
            test: /[\\/]node_modules[\\/](zxcvbn)[\\/]/,
            chunks: 'all'
          }
        },
        name: false
      }
    }
  },
  chainWebpack: config => {
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
