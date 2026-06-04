# Database Testing Approaches with Testfixtures

This document describes different approaches of testing with a main focus on correctness, performance and parallel
execution.

## 1. Sequential single-threaded tests [example_sequential_test.go](example_sequential_test.go)

Use a single shared database connection for all tests. Store your connection and `fixtures` object in a global
variable, so each test can use it.

### Pros:

* 🟢 Simple, set up your database connection once, then use it in all tests
* 🟢 You can use the same `fixtures` object instance for all tests, which speeds up the subsequent uses of
  `fixtures.Load()`
* 🟢 Works well with any database engine

### Cons:

* 🔴 No parallelization, huge performance issue for a project with a reasonable number of tests
* 🔴 Global dependency is usually not a pretty thing to do
* 🔴 You need to prevent test execution from parallel execution:
  * use `go test -p 1 ./...` to test only one package at a given time, if you have database tests in multiple packages
  * do not use `t.Parallel()` in your database tests

## 2. Separate Database Per Test [example_separate_database_per_test.go](example_separate_database_per_test.go)

Create a disposable database for each test.

Both [github.com/testcontainers/testcontainers-go](https://github.com/testcontainers/testcontainers-go)
and [github.com/ory/dockertest](https://github.com/ory/dockertest) are good solutions, which uses a `docker` under-the-hood to create a fresh container for each test.

### Pros:

* 🟢 Perfect isolation
* 🟢 Good for parallel execution
* 🟢 Works well with any database engine; we test all supported engines in CI using this approach

### Cons:

* 🔴 Database setup is usually very slow.
  * some like `postgres` boots fast in less that `1s`
  * some may take `10s` or more
* 🔴 Requires `docker` runtime or any other way to run containers

## 3. Run each test in a transaction [example_txdb_test.go](example_txdb_test.go)

Use a single database. Each test starts the new transaction, which is then `ROLLBACK` after the test completes.
Transactions keep the tests from interfering with each other; although the isolation may not be perfect.

You can do the transaction management manually or use a library
like [github.com/DATA-DOG/go-txdb](https://github.com/DATA-DOG/go-txdb), which
wraps the `*sql.DB` interface and provides a transaction manager on top of it

### Pros:

* 🟢 Superfast, probably the fastest way, when applied in the right context
* 🟢 Good for parallel execution
* 🟢 Easy to set up

### Cons:

* 🔴 May not work for all database engines. `go-txdb` supports only `postgres` and `mysql` at the moment
* 🔴 Transaction may not isolate tests from each other in all cases like DDL operations, which `testfixtures` uses
  heavily
* 🔴 DDL operations and others may slow/lock the database. In worst scenario the whole test suite may run as slow as
  sequential tests
* 🔴 May not work for all operations like `transaction in transaction`

## 4. Create a fresh logical database from a template [example_pgtestdb_test.go](example_pgtestdb_test.go)

You can create your database template once, then each test creates a new logical database from the template.
The example uses [github.com/peterldowns/pgtestdb](https://github.com/peterldowns/pgtestdb) as a helper, but it can be
done also manually
using database-specific operations.

For example, in `postgres` you can use:

```sql
-- create and prepare a template database
CREATE DATABASE dbname TEMPLATE template;
-- use your dbname database during test
```

### Pros:

* 🟢 Almost perfect isolation: the physical instance is shared, but it is almost never a problem
* 🟢 Fast assuming the template clone is fast.
* 🟢 Good for parallel execution

### Cons:

* 🔴 Database clone may be slow; it depends on the complexity of a template database
* 🔴 Requires some setup: prepare a template, create a test database from template, clean the test database
* 🔴 Database specific library/approach is required

