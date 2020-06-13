# widgets

## Project setup
```
yarn install
```

### Compiles and hot-reloads for development
```
yarn run serve
```

### Compiles and minifies for production
```
yarn run build
```

### Setting up the access token

Please refer to the project [README.md](../README.md)

### Run your tests
```
yarn test
```

### Lints and fixes files
```
yarn run lint
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

### Usage note
For inline widget, if the widget is used within a flexbox element(i.e. Parent element with style `display: flex;`), the parent element should set style `overflow: auto;` (or `min-height: 0;`, in case `overflow: visible;` is set) to ensure the element can be scrolled.

### Miscellaneous note

Save password feature
---
For browser save password feature, different browsers have its corresponding behaviour:
- Firefox
  - Get first input field before field with `type="password"`
- Safari
  - Save password feature will only be triggered in non-localhost domain
  - Save password alert is blocking when triggered
- Chrome
  - Save password window will not be triggered if it has been closed for several times(Need further check)

Enable Webpack bundle analyzer in development
---
Since the project using docker and vue-cli-service, the common approach apply webpack-bundle-analyzer on `vue.config.js` can only show stat size. To show all size from the webpack-bundle-analyzer, it is required to build the file and generate report file, using webpack-bundle-analyzer in command line to show the result.

To use webpack bundle analyzer, uncomment port 8888 line in `docker-compose.yaml` file.

In the container env, using the following command to build and run webpack-bundle-analyzer

```sh
# Build the files in production mode, this fits what happens in production environment
./node_modules/.bin/vue-cli-service build --report-json

# Run webpack-bundle-analyzer and check the json from buliding process. Host param is required as by default it is hosted in 127.0.0.1 which cannot be accessed outside container
./node_modules/.bin/webpack-bundle-analyzer ./dist/widgets/report.json --host 0.0.0.0
```
