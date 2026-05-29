# cloudeventscheduler Helm chart

## Install

```sh
helm install cloudeventscheduler oci://ghcr.io/a5r0n/charts/cloudeventscheduler \
  --version 0.2.0 \
  --set config.database.dsn="postgres://user:pass@host:5432/dbname?sslmode=require" \
  --set config.redis="redis://:pass@host:6379/0"
```

## External Postgres + Redis are mandatory

As of chart `0.2.0` this chart no longer bundles Redis or Postgres
subcharts. The previous `bitnami/redis` and `bitnami/postgresql`
dependencies were removed because the Bitnami public Helm catalog
(`charts.bitnami.com/bitnami`) was deleted on 2025-09-29, making
`helm dependency build` fail and the chart unusable for fresh installs.

You must provide connection strings for both:

| Value                  | Required | Description                                |
| ---------------------- | -------- | ------------------------------------------ |
| `config.database.dsn`  | yes      | Postgres DSN, e.g. `postgres://...`        |
| `config.redis`         | yes      | Redis URL, e.g. `redis://...` or `rediss://...` |

Suggested replacements for the removed subcharts:

- Redis: `valkey/valkey-bundle` (LF-stewarded Valkey fork) or a managed
  Redis-compatible service.
- Postgres: CloudNativePG, Zalando postgres-operator, or a managed
  Postgres service.

## OCI registry

Releases are published as OCI artifacts:

```
oci://ghcr.io/a5r0n/charts/cloudeventscheduler
```
