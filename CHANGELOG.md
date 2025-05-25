# Changelog

## v3.16.0 - 2025-05-25
- feat: migrate from gopkg.in/yaml.v3 to github.com/goccy/go-yaml ([#290](https://github.com/go-testfixtures/testfixtures/pull/290) by @tomnewton)
- feat: add support for interleaved tables in spanner ([#290](https://github.com/go-testfixtures/testfixtures/pull/290) by @tomnewton)
- perf: identify identity columns once during init for postgres (#289) ([#289](https://github.com/go-testfixtures/testfixtures/pull/289) by @kolaente)
- feat: adds support for foreign keys constraints for composite primary keys in spanner ([#287](https://github.com/go-testfixtures/testfixtures/pull/287) by @tomnewton)

## v3.15.0 - 2025-05-11

- feat: support quoted columns in postgresql ([#286](https://github.com/go-testfixtures/testfixtures/pull/286) by @HTechHQ)
- test: move all db tests to dbtest package ([#251](https://github.com/go-testfixtures/testfixtures/pull/251) by @slsyy)
- chore: remove spanner underscore import ([#250](https://github.com/go-testfixtures/testfixtures/pull/250) by @slsyy)
- Upgraded to Go v1.23 ([#281](https://github.com/go-testfixtures/testfixtures/pull/281) by @slsyy)
- Updated golangci-lint and CI ([#281](https://github.com/go-testfixtures/testfixtures/pull/281) by @slsyy)
- Updated dependencies.

## v3.14.0 - 2024-12-22

- feat(mysql): make multistatements parameter optional ([#249](https://github.com/go-testfixtures/testfixtures/pull/249) by @slsyy)
- test: remove private api usage in assertFixturesLoaded ([#239](https://github.com/go-testfixtures/testfixtures/pull/239) by @slsyy)
- Updated dependencies.

## v3.13.0 - 2024-10-25

- Add GCP Spanner support ([#211](https://github.com/go-testfixtures/testfixtures/pull/211) by @kikihakiem)
- Remove ClickHouse underscore import by ([#220](https://github.com/go-testfixtures/testfixtures/pull/220) by @slsyy)
- test: remove private api usages in tests ([#221](https://github.com/go-testfixtures/testfixtures/pull/221) by @slsyy)
- CI: use `docker compose` instead of `docker-compose` ([#214](https://github.com/go-testfixtures/testfixtures/pull/214) by @slsyy)
- Updated dependencies.

## v3.12.0 - 2024-07-13

- Reset sequences in a single exec to improve performance
  ([#208](https://github.com/go-testfixtures/testfixtures/pull/208) by @slsyy)
- Skip checksum calculation when not needed to improve performance
  ([#207](https://github.com/go-testfixtures/testfixtures/pull/207) by @slsyy).
- Add `SkipTableChecksumComputation` option
  ([#203](https://github.com/go-testfixtures/testfixtures/issues/203), [#206](https://github.com/go-testfixtures/testfixtures/pull/206) by @slsyy)
- PostgreSQL: Run some queries concurrently to improve performance
  ([#205](https://github.com/go-testfixtures/testfixtures/pull/205) by @slsyy).
- Optimize Docker image a bit
  ([#204](https://github.com/go-testfixtures/testfixtures/pull/204) by @slsyy).

## v3.11.0 - 2024-05-25

- Add `OVERRIDING SYSTEM VALUE` for `INSERT` statements on PostgreSQL
  ([#183](https://github.com/go-testfixtures/testfixtures/pull/183) by @amakmurr).
- Upgraded dependencies.

## v3.10.0 - 2024-02-17

- Fix usage with Microsoft SQL Server when the database is configured with a
  case sensitive setting ([#182](https://github.com/go-testfixtures/testfixtures/pull/182) by @wxiaoguang).
- Updated some dependencies.
- Updated database systems versions on the Docker setup used by CI
  ([#187](https://github.com/go-testfixtures/testfixtures/pull/187)).

## v3.9.0 - 2023-05-01

- Added support do ClickHouse
  ([#51](https://github.com/go-testfixtures/testfixtures/issues/51), [#162](https://github.com/go-testfixtures/testfixtures/pull/162) by @titusjaka, [#115](https://github.com/go-testfixtures/testfixtures/pull/115) by @shumorkiniv, [#81](https://github.com/go-testfixtures/testfixtures/pull/81) by @kangoo13).
- Add option to disable database cleanup
  ([#161](https://github.com/go-testfixtures/testfixtures/pull/161)).
- Start releasing binaries for Mac M1
  ([#149](https://github.com/go-testfixtures/testfixtures/issues/149), [#150](https://github.com/go-testfixtures/testfixtures/pull/150)).
- Upgraded to Go v1.20
  ([#165](https://github.com/go-testfixtures/testfixtures/pull/165)).
- Upgraded several dependencies.

## v3.8.1 - 2022-08-01

- Upgrade `golang.org/x/crypto` dependency that includes a security fix
  ([#136](https://github.com/go-testfixtures/testfixtures/pull/136)).

## v3.8.0 - 2022-07-04

- Add ability to load from a custom filesystem
  ([#134](https://github.com/go-testfixtures/testfixtures/pull/134), [pkg.go.dev/io/fs](https://pkg.go.dev/io/fs)).
- Upgrade to gopkg.in/yaml.v3, which includes a possible security vulnerability
  ([#132](https://github.com/go-testfixtures/testfixtures/pull/132), [go-yaml/yaml#666](https://github.com/go-yaml/yaml/issues/666)).

## v3.7.0 - 2022-05-29

- Add support for declaring multiples tables in the same YAML file
  ([#98](https://github.com/go-testfixtures/testfixtures/issues/98), [#130](https://github.com/go-testfixtures/testfixtures/pull/130)).
- Upgrade dependencies

## v3.6.2 - 2022-05-15

- Upgrade dependencies

## v3.6.1 - 2021-05-20

- Fix possible security vulnerability by upgrading golang.org/x/crypto
  ([#100](https://github.com/go-testfixtures/testfixtures/pull/100)).

## v3.6.0 - 2021-04-17

- Add support for dumping a database using the CLI (use the `--dump` flag)
  ([#88](https://github.com/go-testfixtures/testfixtures/pull/88), [#63](https://github.com/go-testfixtures/testfixtures/issues/63)).
- Support SkipResetSequences and ResetSequencesTo for MySQL and MariaDB
  ([#91](https://github.com/go-testfixtures/testfixtures/pull/91)).

## v3.5.0 - 2021-01-11

- Fix insert of JSON values on PostgreSQL when using `binary_parameters=yes` in
  the connection string
  ([#83](https://github.com/go-testfixtures/testfixtures/issues/83), [#84](https://github.com/go-testfixtures/testfixtures/pull/84), [lib/pq#528](https://github.com/lib/pq/issues/528)).
- Officially support binary columns through hexadecimal strings
  ([#48](https://github.com/go-testfixtures/testfixtures/issues/48), [#82](https://github.com/go-testfixtures/testfixtures/pull/82)).

## v3.4.1 - 2020-10-19

- Fix for Microsoft SQL Server databases with views
  ([#78](https://github.com/go-testfixtures/testfixtures/pull/78)).

## v3.4.0 - 2020-08-09

- Add support to CockroachDB
  ([#77](https://github.com/go-testfixtures/testfixtures/pull/77)).

## v3.3.0 - 2020-06-27

- Add support for the [github.com/jackc/pgx](https://github.com/jackc/pgx)
  PostgreSQL driver
  ([#71](https://github.com/go-testfixtures/testfixtures/issues/71), [#74](https://github.com/go-testfixtures/testfixtures/pull/74)).
- Fix bug where some tables were empty due to `ON DELETE CASCADE`
  ([#67](https://github.com/go-testfixtures/testfixtures/issues/67), [#70](https://github.com/go-testfixtures/testfixtures/pull/70)).
- Fix SQLite version
  ([#73](https://github.com/go-testfixtures/testfixtures/pull/73)).
- On MySQL, return a clearer error message when a table doesn't exist
  ([#69](https://github.com/go-testfixtures/testfixtures/pull/69)).

## v3.2.0 - 2020-05-10

- Add support for loading multiple files and directories
  ([#65](https://github.com/go-testfixtures/testfixtures/pull/65)).

## v3.1.2 - 2020-04-26

- Dump: Fix column order in generated YAML files
  ([#62](https://github.com/go-testfixtures/testfixtures/pull/62)).

## v3.1.1 - 2020-01-11

- testfixtures now work with both `mssql` and `sqlserver` drivers.
  Note that [the `mssql` one is deprecated](https://github.com/denisenkom/go-mssqldb#deprecated),
  though. So try to migrate to `sqlserver` once possible.

## v3.1.0 - 2020-01-09

- Using `sqlserver` driver instead of the deprecated `mssql`
  ([#58](https://github.com/go-testfixtures/testfixtures/pull/58)).

## v3.0.0 - 2019-12-26

### Breaking changes

- The import path changed from `gopkg.in/testfixtures.v2` to
  `github.com/go-testfixtures/testfixtures/v3`.
- This package no longer support Oracle databases. This decision was
  taken because too few people actually used this package with Oracle and it
  was the most difficult to test (we didn't run on CI due the lack of an
  official Docker image, etc).
- The public API was totally rewritten to be more flexible and ideomatic.
  It now uses functional options. It differs from v2, but should be easy
  enough to upgrade.
- Some deprecated APIs from v2 were removed as well.
- This now requires Go >= 1.13.

### New features

- We now have a CLI so you can easily use testfixtures to load a sample
  database from fixtures if you want.
- Templating via [text/template](https://golang.org/pkg/text/template/)
  is now available. This allows some fancier use cases like generating data
  or specific columns dynamically.
- It's now possible to choose which time zone to use when parsing timestamps
  from fixtures. The default is the same as before, whatever is set on
  `time.Local`.
- Errors now use the new `%w` verb only available on Go >= 1.13.

### MISC

- Travis and AppVeyor are gone. We're using GitHub Actions exclusively now.
  The whole suite is ran inside Docker (with help of Docker Compose), so it's
  easy to run tests locally as well.

Check the new README for some examples!

## v2.6.0 - 2019-10-24

- Add support for TimescaleDB
  ([#53](https://github.com/go-testfixtures/testfixtures/pull/53)).

## v2.5.3 - 2018-12-15

- Fixes related to use of foreign key pragmas on MySQL (#43).

## v2.5.2 - 2018-11-25

- This library now supports [Go Modules](https://github.com/golang/go/wiki/Modules);
- Also allow `.yaml` (as an alternative to `.yml`) as the file extension (#42).

## v2.5.1 - 2018-11-04

- Allowing disabling reset of PostgreSQL sequences (#38).

## v2.5.0 - 2018-09-07

- Add public function DetectTestDatabase (#35, #36).

## v2.4.5 - 2018-07-07

- Fix for MySQL/MariaDB: ignoring views on operations that should be run only on tables (#33).

## v2.4.4 - 2018-07-02

- Fix for multiple schemas on Microsoft SQL Server (#29 and #30);
- Configuring AppVeyor CI to also test for Microsoft SQL Server.

---

Sorry, we don't have changelog for older releases 😢.
