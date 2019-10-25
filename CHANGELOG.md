# Changelog

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
