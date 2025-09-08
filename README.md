## Warning - Unstable and under development

## This project is not affiliated nor supported in any way, shape, or form with MuleSoft, LLC or Salesforce

# Anypoint Monitoring Datasource

The Anypoint Monitoring Datasource is developed in house by City Geo, City Of Philadelphia as a simple way to extract metrics from Anypoint Monitoring into Grafana for centralized monitoring and alerting. This project was quickly put together and is currently lacking many best practices.

Currently lacking features:

- Any type of CI/CD for automated testing
- Any type of variable discovery in Grafana (it does not help you figure out what variables, tables, etc. are available)
- The ability to build custom queries
- Only a very specific metrics monitoring endpoint is available
- Proper error messages and logging

## Features

- Queries Anypoint Monitoring InfluxDB and retrieves basic metrics
- Handles access token creation through Client ID and Client Secret

## Where it gets Data from

While Anypoint monitoring has an official [Metrics API](https://anypoint.mulesoft.com/exchange/portals/anypoint-platform/f1e97bc6-315a-4490-82a7-23abe036327a.anypoint-platform/metrics-api/), it is lacking several crucial metrics, most importantly the JVM metrics which were the main driver behind this project.

Therefore, this project pulls directly from the internal InfluxDB API used in Anypoint Monitoring. The link to that URL looks like this: <https://anypoint.mulesoft.com/monitoring/api/visualizer/api/datasources/proxy/148362>. To better understand how the queries work, it is recommended to go to Anypoint Monitoring in your browser and view the XHR requests it makes to this path.

## The current query

The current query made to that endpoint is available in the "queryStringTemplateRaw" variable in the "NewDatasource" function in "pkg/plugin/datasource.go". As of 2025-09-08, it is:

```sql
SELECT mean("{{ .metric_name }}") FROM "{{ .metric_table }}" WHERE ("org_id" = '{{ .org_id }}' AND "env_id" = '{{ .env_id }}' AND "cluster_id" = '{{ .cluster_id }}' AND "app_id" = '{{ .app_id }}') AND time >= {{ .start_time }}ms and time <= {{ .end_time }}ms GROUP BY time({{ .time_step }}), "worker_id" fill(null) tz('America/New_York')
```

## Development

This project has only been developed enough to a point where the City of Philadelphia can successfully use it to monitor and alert on their own metrics. If there is interest from other groups or organizations to further develop this plugin, please contact Ryan Weast (<ryweast@gmail.com>).

### Development recommendations

Since MuleSoft is just using InfluxDB on the backend, perhaps it is possible to just fork the existing InfluxDB Grafana data source and configure it to work.

---

# Grafana data source plugin template

This template is a starting point for building a Data Source Plugin for Grafana.

## What are Grafana data source plugins?

Grafana supports a wide range of data sources, including Prometheus, MySQL, and even Datadog. There’s a good chance you can already visualize metrics from the systems you have set up. In some cases, though, you already have an in-house metrics solution that you’d like to add to your Grafana dashboards. Grafana Data Source Plugins enables integrating such solutions with Grafana.

## Getting started

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/developers/plugin-tools/key-concepts/backend-plugins/grafana-plugin-sdk-for-go) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build plugin backend binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```

### Frontend

1. Install dependencies

   ```bash
   npm install
   ```

2. Build plugin in development mode and run in watch mode

   ```bash
   npm run dev
   ```

3. Build plugin in production mode

   ```bash
   npm run build
   ```

4. Run the tests (using Jest)

   ```bash
   # Runs the tests and watches for changes, requires git init first
   npm run test

   # Exits after running all the tests
   npm run test:ci
   ```

5. Spin up a Grafana instance and run the plugin inside it (using Docker)

   ```bash
   npm run server
   ```

6. Run the E2E tests (using Playwright)

   ```bash
   # Spins up a Grafana instance first that we tests against
   npm run server

   # If you wish to start a certain Grafana version. If not specified will use latest by default
   GRAFANA_VERSION=11.3.0 npm run server

   # Starts the tests
   npm run e2e
   ```

7. Run the linter

   ```bash
   npm run lint

   # or

   npm run lint:fix
   ```

# Distributing your plugin

When distributing a Grafana plugin either within the community or privately the plugin must be signed so the Grafana application can verify its authenticity. This can be done with the `@grafana/sign-plugin` package.

_Note: It's not necessary to sign a plugin during development. The docker development environment that is scaffolded with `@grafana/create-plugin` caters for running the plugin without a signature._

## Initial steps

Before signing a plugin please read the Grafana [plugin publishing and signing criteria](https://grafana.com/legal/plugins/#plugin-publishing-and-signing-criteria) documentation carefully.

`@grafana/create-plugin` has added the necessary commands and workflows to make signing and distributing a plugin via the grafana plugins catalog as straightforward as possible.

Before signing a plugin for the first time please consult the Grafana [plugin signature levels](https://grafana.com/legal/plugins/#what-are-the-different-classifications-of-plugins) documentation to understand the differences between the types of signature level.

1. Create a [Grafana Cloud account](https://grafana.com/signup).
2. Make sure that the first part of the plugin ID matches the slug of your Grafana Cloud account.
   - _You can find the plugin ID in the `plugin.json` file inside your plugin directory. For example, if your account slug is `acmecorp`, you need to prefix the plugin ID with `acmecorp-`._
3. Create a Grafana Cloud API key with the `PluginPublisher` role.
4. Keep a record of this API key as it will be required for signing a plugin

## Signing a plugin

### Using Github actions release workflow

If the plugin is using the github actions supplied with `@grafana/create-plugin` signing a plugin is included out of the box. The [release workflow](./.github/workflows/release.yml) can prepare everything to make submitting your plugin to Grafana as easy as possible. Before being able to sign the plugin however a secret needs adding to the Github repository.

1. Please navigate to "settings > secrets > actions" within your repo to create secrets.
2. Click "New repository secret"
3. Name the secret "GRAFANA_API_KEY"
4. Paste your Grafana Cloud API key in the Secret field
5. Click "Add secret"

#### Push a version tag

To trigger the workflow we need to push a version tag to github. This can be achieved with the following steps:

1. Run `npm version <major|minor|patch>`
2. Run `git push origin main --follow-tags`

## Learn more

Below you can find source code for existing app plugins and other related documentation.

- [Basic data source plugin example](https://github.com/grafana/grafana-plugin-examples/tree/master/examples/datasource-basic#readme)
- [`plugin.json` documentation](https://grafana.com/developers/plugin-tools/reference/plugin-json)
- [How to sign a plugin?](https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin)
