# JDBC Driver Management Design

Date: 2026-07-09

## Background

The current database module connects through Go `database/sql` and registers only MySQL and PostgreSQL drivers. Adding every known Go database driver to the main app would increase binary size, force releases for driver changes, and still leave weak coverage for enterprise and domestic database vendors.

The target direction is an all-JDBC architecture: the Go/Wails app owns product state, credentials, installation, and process lifecycle; a Java JDBC agent owns driver loading, database sessions, SQL execution, and metadata access. This gives the product a DBeaver/DBX-like driver model where drivers and Java runtime can be installed only when needed.

## Goals

- Support many database types without compiling every driver into the Go binary.
- Use JDBC as the only database execution path for the database module.
- Provide first-class online and offline installation flows for JRE and JDBC drivers.
- Keep ordinary connection creation simple: users choose database type, not driver internals.
- Preserve advanced overrides for driver class, JDBC URL template, version, Maven coordinates, and extra jars.
- Keep the existing Wails API surface conceptually stable for frontend database panels.

## Non-Goals

- Data editing, import/export, schema diff, ER diagrams, and migration tooling are outside the first version.
- Go-native database drivers are not part of this design, except as existing code to be replaced or bypassed.
- Public network access to the JDBC agent is not supported.

## Recommended Architecture

The application should use an all-JDBC agent model.

```text
Svelte UI
  -> Wails App API
  -> Go JdbcGatewayService
  -> Go AgentProcessManager
  -> Java JDBC Agent
  -> JDBC Driver jar
  -> Database
```

The Go app remains the product shell and source of truth for settings. The Java agent is a local worker process launched by Go. Go and the Java agent communicate through gRPC bound to `127.0.0.1`.

Runtime strategy:

- Default: automatically download and install a managed JRE on first use.
- Offline: import a JRE package from disk.
- Advanced: point to a system Java installation.
- Lifecycle: start the agent on demand, allow a setting to keep it resident, and shut it down after an idle timeout.

Driver strategy:

- Official driver manifest is the primary source.
- Custom Maven coordinates and manually imported jars are advanced options.
- All official downloads and offline packages require checksum validation.
- Drivers and runtimes are stored under `~/.sshtools/`.

Initial JDBC database set:

- MySQL
- PostgreSQL
- SQLite
- Oracle
- SQL Server
- DM
- Kingbase
- openGauss

Initial database capabilities:

- Test connection
- Open and close session
- Execute SQL
- Return query result columns, rows, and affected count
- List catalogs/databases, schemas, tables, and columns

## Filesystem Layout

```text
~/.sshtools/
  config.json
  credentials.enc
  runtimes/
    jre-21-<os>-<arch>/
  drivers/
    manifest.json
    mysql/
      8.4.0/
        mysql-connector-j.jar
        driver.json
    oracle/
      23.x/
        ojdbc11.jar
        driver.json
  agent/
    jdbc-agent.jar
  logs/
    jdbc-agent.log
    driver-install.log
    runtime-install.log
```

## Driver Package Format

Offline driver packages use a zip layout:

```text
driver-package.zip
  package.json
  jars/*.jar
  checksums.sha256
  licenses/*
```

Example `package.json`:

```json
{
  "id": "oracle",
  "name": "Oracle",
  "version": "23.5",
  "driverClass": "oracle.jdbc.OracleDriver",
  "urlTemplate": "jdbc:oracle:thin:@//{host}:{port}/{database}",
  "defaultPort": 1521,
  "jre": ">=17",
  "jars": ["ojdbc11.jar"]
}
```

## Go Module Boundaries

`DriverCatalogService`

- Reads official manifest.
- Reads local installed manifest.
- Reads user-defined custom profiles.
- Answers which drivers exist, which versions are recommended, and which are installed.

`DriverInstallService`

- Downloads official driver packages.
- Imports offline packages.
- Validates checksums.
- Installs, uninstalls, exports, and rolls back driver files.
- Emits install task status.

`RuntimeService`

- Detects managed JRE, imported JRE, and system Java.
- Selects active runtime.
- Installs and uninstalls managed runtime files.
- Validates Java version compatibility.

`AgentProcessManager`

- Starts the Java agent on demand.
- Passes local gRPC port and token.
- Performs health checks.
- Restarts or stops the agent.
- Applies idle shutdown and resident mode settings.

`JdbcGatewayService`

- Keeps the Wails-facing database API stable.
- Converts connect, query, metadata, and close calls into gRPC requests.
- Maps gRPC/agent errors into UI-friendly database error categories.

`DatabaseProfileService`

- Renders JDBC URL templates.
- Supplies default ports, driver classes, and connection properties.
- Supports advanced user overrides.

## Java Agent Module Boundaries

`DriverLoader`

- Creates an isolated classloader per driver profile.
- Loads JDBC driver jars.
- Avoids dependency conflicts between vendors.

`ConnectionRegistry`

- Maps `sessionID` to JDBC `Connection`.
- Owns connection close and cleanup.

`QueryService`

- Executes SQL.
- Distinguishes result-set queries from update statements.
- Returns columns, rows, and affected count.

`MetadataService`

- Uses `DatabaseMetaData` to list catalogs, schemas, tables, and columns.
- Allows database-specific adjustments where JDBC metadata is incomplete.

`HealthService`

- Reports agent version, runtime version, installed driver visibility, and active sessions.

## Driver Management UI

Use a two-column driver manager page.

Left column:

- Search input.
- Filters: installed, available, update available, offline packages, custom JDBC.
- Database list with compact status badges: not installed, installed, update available, validation failed, missing dependency.

Right column:

- Header with database name, recommended version, install status, license, package size, and source.
- Version section with recommended version, installed version, historical versions, switch, and rollback.
- File section with jars, checksums, install path, and dependency jars.
- Advanced section with driver class, URL template, default port, connection properties, Maven coordinates, and extra jars.
- Actions: install, reinstall, uninstall, import offline package, export offline package, validate, test driver.
- Bottom task strip showing download, extract, checksum, install, retry, and failure state.

JRE management should appear in the same driver manager area, either as a top band or a settings subsection. It includes managed JRE, system Java, offline import, version detection, switch, and uninstall.

Primary user path:

```text
Open Driver Manager -> select database -> install recommended driver -> create connection -> test -> connect
```

Advanced user path:

```text
Open Driver Manager -> add custom JDBC profile -> import jars or Maven coordinates -> validate -> create connection with profile override
```

## Connection Configuration

Connection profiles should store:

- `db_type`
- host
- port
- database/service name
- user
- encrypted password reference
- optional `driver_profile_id`
- optional connection properties

Ordinary users choose only the database type. The app binds each database type to a recommended driver profile. Advanced settings allow users to override profile, version, class, URL template, and jars.

## Error Handling

Errors should be categorized and paired with actionable UI buttons.

- `RUNTIME_MISSING`: no valid JRE. Actions: install managed JRE, import offline JRE, choose system Java.
- `DRIVER_MISSING`: driver not installed. Actions: install recommended driver, import offline package.
- `DRIVER_INVALID`: missing jar, checksum mismatch, class loading failure. Actions: reinstall, view files, delete.
- `AGENT_UNAVAILABLE`: startup failure, gRPC unavailable, version mismatch. Actions: restart agent, view logs.
- `DB_CONNECT_FAILED`: authentication, network, URL, or database rejection. Actions: edit connection, test network, view original error.

SQL errors should preserve vendor messages, but passwords, tokens, and full JDBC URLs must be redacted.

## Security

- The agent listens only on `127.0.0.1`.
- Go generates a per-agent startup token; every gRPC request must include it.
- The Java agent does not read `credentials.enc`, SSH keys, or app config directly.
- Go decrypts credentials and sends only the needed password to the agent for session creation.
- Passwords live only in memory inside the agent session.
- Official package installs require checksum validation.
- Offline packages require checksum validation.
- Custom Maven coordinates are marked as unverified and should not auto-update by default.
- Uninstalling a driver checks whether connection profiles depend on it and shows the impact list.

## Testing Strategy

Go unit tests:

- Manifest parsing.
- Version selection.
- Checksum validation.
- Driver path calculation.
- Offline package import rollback.
- Runtime selection priority.
- Agent start, health check, idle stop, and restart behavior.
- gRPC error mapping in `JdbcGatewayService`.

Java agent unit tests:

- Classloader isolation.
- JDBC URL template rendering.
- Query result mapping for select and update statements.
- Metadata listing using H2 or SQLite JDBC.

Integration tests:

- Use H2 or SQLite JDBC as no-external-service test database.
- Start Java agent from Go.
- Install local test driver package.
- Connect, query, list tables, list columns, and close session.
- Kill agent and confirm Go marks sessions disconnected and reports a reconnect action.

Manual acceptance:

- First database use on a machine without Java prompts managed JRE install.
- Offline machine can import JRE package and driver package.
- Oracle, SQL Server, DM, Kingbase, and openGauss can run connection test and simple query.

## Rollout Plan

1. Introduce driver/runtime metadata models and local manifest storage.
2. Build Java agent with gRPC health, connect, query, metadata, and close methods.
3. Add Go agent process manager and gateway service.
4. Replace current Go `database/sql` database calls with gateway calls.
5. Add driver manager UI and JRE management.
6. Add offline import/export flows.
7. Validate against the initial database set.

## Residual Risks

- JDBC metadata behavior varies by vendor and may need per-database adapters.
- Some vendor drivers have license restrictions that prevent bundled redistribution.
- Java runtime downloads introduce supply-chain and mirror availability concerns.
- gRPC process management adds cross-platform lifecycle complexity.
- Agent crashes invalidate active sessions; users must reconnect.
