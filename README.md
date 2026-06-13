# rollops-plugin-split

A [Rollops](https://github.com/klarlabs-studio/rollops) feature-flag provider
plugin backed by [Split](https://www.split.io/) (Harness FME). It drives a
split's default rule treatment buckets so the on/off percentage split matches a
Rollops canary in lockstep — as a rollout steps 10% → 50% → 100%, the `on`
bucket follows.

## How it works

Rollops calls the plugin per progressive step (and/or on promote) with the split
name, target environment, and current traffic percentage. The plugin PUTs the
split definition for that environment with a default rule whose treatment buckets
are `on: <pct>` / `off: <100-pct>`. When the change is disabled (rollback), the
`on` bucket is 0 so every key serves `off`.

## Configuration

Credentials come from the plugin's own environment, never from the Rollops
target spec:

| Env var         | Required | Default                | Description           |
|-----------------|----------|------------------------|-----------------------|
| `SPLIT_API_URL` | no       | `https://api.split.io` | Base URL              |
| `SPLIT_TOKEN`   | yes      | —                      | Admin API key (`Bearer`) |

## Install

```sh
rollops plugin install split
```

Or build and pin manually with `make build` / `make checksum`, then wire into a
rollout spec:

```yaml
featureFlags:
  plugin: ~/.rollops/plugins/split
  sha256: <pin>
  flag: checkout
  environment: Production
  when: both
```

## License

MIT
