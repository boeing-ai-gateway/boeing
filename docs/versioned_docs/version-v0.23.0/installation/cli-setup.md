---
title: Boeing CLI Setup
---

The `boeing setup` command prepares your local workstation to use an Boeing server from the command line and from supported local AI clients.

Use it after an Boeing server is running and reachable from your machine.

## What it does

`boeing setup` performs these steps:

1. Resolves the Boeing app URL to use, either from `--url`, from an existing local default, or by prompting you.
2. Authenticates to that Boeing server. If `BOEING_TOKEN` is set, the CLI uses that token. Otherwise, it uses the same browser-based API key flow as `boeing login`.
3. Stores the normalized default Boeing URL in the local Boeing CLI config.
4. Stores a newly acquired Boeing API key in the host OS keyring, scoped to that Boeing URL.
5. Optionally installs Boeing bootstrap skills into supported local AI clients.

The bootstrap skills let local agents use the `boeing` CLI to search for Boeing-managed skills, install skills, and run local client scans without manually editing client configuration.

:::note
`boeing setup` configures the local CLI and local client bootstrap files. It does not deploy the Boeing server or configure server-side authentication providers.
:::

## Prerequisites

- The `boeing` CLI is installed and available on your `PATH`.
- The Boeing server URL is reachable from your workstation.
- If authentication is enabled, Boeing has at least one configured authentication provider that your user can use.
- Your local OS keyring is available so the CLI can store a newly acquired API key.

If Boeing authentication is enabled but no provider is configured yet, finish server-side authentication setup first. See [Enabling Authentication](/installation/enabling-authentication/) and [Auth Providers](/configuration/auth-providers/).

## Basic usage

Run setup with your Boeing app URL:

```bash
boeing setup --url https://boeing.example.com
```

For a local Docker deployment using the default port:

```bash
boeing setup --url http://localhost:8080
```

If authentication is required, the CLI opens a browser to complete login. After login succeeds, setup saves the default URL and asks where to install local bootstrap skills.

## Choosing local client targets

Use `--clients` to choose where bootstrap skills are installed:

| Value | Description | Install location |
|-------|-------------|------------------|
| `agents` | Install into the shared Agent Skills directory used by clients that support `~/.agents`. | `~/.agents/skills` |
| `claude-code` | Install into Claude Code's skills directory. | `~/.claude/skills` |
| `none` | Skip local client bootstrap installation. | Not applicable |

You can install into more than one target:

```bash
boeing setup --url https://boeing.example.com --clients agents,claude-code
```

To configure only CLI authentication and the default URL:

```bash
boeing setup --url https://boeing.example.com --clients none
```

When `--clients` is omitted in an interactive terminal, setup prompts you. The prompt always offers `agents`. It offers `claude-code` when Claude Code is detected locally. You can still install Claude Code support explicitly with `--clients claude-code`.

## Non-interactive setup

For scripts or GUI wrappers, pass both `--url` and `--clients` with `--non-interactive`:

```bash
boeing setup \
  --url https://boeing.example.com \
  --clients agents \
  --non-interactive
```

Non-interactive mode never reads from stdin. It still uses the normal API key flow, so it may open a browser and wait for authentication unless a valid key is already stored.

Use `--yes` to accept defaults and confirmations. If `--clients` is omitted with `--yes`, setup installs the shared `agents` target by default:

```bash
boeing setup --url https://boeing.example.com --yes
```

If a different default Boeing URL is already configured, setup refuses to replace it unless you pass `--yes`:

```bash
boeing setup --url https://new-boeing.example.com --yes
```

## Check setup status

Use `boeing setup status` to verify the local configuration:

```bash
boeing setup status
```

The command prints:

- CLI version
- Default Boeing URL
- Whether the stored API key is valid
- Whether setup is complete

For JSON output:

```bash
boeing setup status --json
```

## What setup writes locally

`boeing setup` writes:

- The default Boeing URL to the Boeing CLI config file under the user's XDG config directory.
- An API key to the host OS keyring under the `boeing` service, scoped by Boeing app URL, when setup acquires a new key through the login flow.
- Bootstrap skill files under the selected client skill directories, such as `~/.agents/skills` or `~/.claude/skills`.

## Troubleshooting

### `auth_unavailable`

The Boeing server did not report exactly one usable configured authentication provider. Configure an auth provider first, or use an interactive setup flow if multiple providers are configured and you need to choose one.

### `server_unreachable`

Check that the URL points to the Boeing app, that the server is running, and that the CLI can reach it from your workstation.

### Missing `--url` in non-interactive mode

Pass `--url`, or run setup interactively and enter the URL when prompted.

### `--clients is required in non-interactive mode`

Pass `--clients agents`, `--clients claude-code`, `--clients agents,claude-code`, or `--clients none`.

### Existing URL mismatch

If setup reports that another Boeing URL is already configured, pass `--yes` to replace the stored default URL.
