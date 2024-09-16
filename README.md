# monarch

monarch is a simple portable database migration tool. You can use it as command line tool or
import/copy it into your Go project.

## Basic usage

### Installation in a Go app

1. Add `github.com/tinyprint/monarch` to your tools file. This will prevent `go mod tidy` from
   removing the dependency and ensure every environment is using the same version of monarch.

   ```go
   //go:build tools
   // +build tools

   package tools

   import (
       _ "github.com/tinyprint/monarch"
   )
   ```

2. Run `go get github.com/tinyprint/monarch` to install monarch
3. Setup configuration environment variables:

   ```env
   DATABASE_URL=postgres://username:password@127.0.0.1:5432/database_name
   MIGRATIONS_PATH=./migrations
   ```

   On development, you can place these environment variables in a `.env` file and run:

   ```sh
   source .env
   export DATABASE_URL MIGRATIONS_PATH
   ```

   On production, you should set these environments variables in however you normally secure secret
   environment variables.

4. Run `go run github.com/tinyprint/monarch init` to initialize monarch. This will create your
   migrations directory and create a `template.lua.tmpl` file that all future migrations will be used
   as a template when creating new migrations.
5. Run `go run github.com/tinyprint/monarch create your_migration_name`. This will create a new
   migration file where you can build out your migration script.
6. Run `go run github.com/tinyprint/monarch migrate`. This will run any migrations that have not run yet.

## Design decisions

- **Lua is used to write migrations.** Monarch intentionally uses a scripting language that is
  normally used for plug-ins to minimize the likelihood that application code can be referenced from
  migrations. Using a migration tool that matches the language of your application makes it too easy
  for engineers to want to reuse application code in migrations. Relying on application code in
  migrations leads to needing to update migrations when application code changes, and migrations
  should never change after being deployed to production.
