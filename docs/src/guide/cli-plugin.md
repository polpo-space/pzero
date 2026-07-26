---
title: Custom CLI plugins
icon: /icons/arcticons-game-plugins.svg
star: true
order: 0.15
---

If built-in commands are not enough, you can extend `pzero` with external executables instead of modifying the main binary.

This is useful for team-specific scaffolding, internal release workflows, deployment helpers, or any command that should feel like a native `pzero` subcommand.

## Discovery rules

When `pzero` receives an unknown command, it searches `PATH` for matching plugin executables:

* `pzero hello` -> `pzero-hello`
* `pzero foo bar` -> first tries `pzero-foo-bar`, then falls back to `pzero-foo`
* After a plugin is matched, the remaining arguments are passed through to the plugin unchanged
* The current environment variables are also forwarded to the plugin process

A plugin only needs two requirements:

* The file name starts with `pzero-`
* The file is executable and available in `PATH`

:::tip
Put plugin-specific flags after the plugin command, for example `pzero hello --name codex`.
:::

## Minimal example

The plugin can be written in Go, shell, or any language that can produce an executable in `PATH`.

```bash
mkdir -p ~/.local/bin

cat > ~/.local/bin/pzero-hello <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

name="${1:-world}"
printf 'hello, %s\n' "$name"
EOF

chmod +x ~/.local/bin/pzero-hello
export PATH="$HOME/.local/bin:$PATH"

pzero hello codex
# hello, codex
```

## Read `desc` metadata in Go plugins

If your plugin is implemented in Go, you can also reuse `github.com/polpo-space/pzero/cmd/pzero/pkg/plugin`.

This does not replace external plugin discovery. `pzero` still discovers your binary through the `pzero-*` naming rule. The extra package is for reading parsed project metadata inside the plugin process.

`plugin.New()` scans the current working directory and attempts to parse:

* `desc/api`
* `desc/proto`
* `desc/sql`

It returns a `Metadata` value whose `Desc` field contains:

* `Desc.Api.SpecMap`: parsed API specs keyed by source file path
* `Desc.Proto.SpecMap`: parsed Proto specs keyed by source file path
* `Desc.Model.SpecMap`: parsed SQL table specs keyed by table name

```go
package main

import (
	"fmt"

	jplugin "github.com/polpo-space/pzero/cmd/pzero/pkg/plugin"
)

func main() {
	metadata, err := jplugin.New()
	if err != nil {
		panic(err)
	}

	fmt.Printf("api files: %d\n", len(metadata.Desc.Api.SpecMap))
	fmt.Printf("proto files: %d\n", len(metadata.Desc.Proto.SpecMap))
	fmt.Printf("sql tables: %d\n", len(metadata.Desc.Model.SpecMap))
}
```

:::tip
`plugin.New()` reads from the plugin process's current working directory, so it is typically used when your plugin is executed inside a pzero project root.
:::

## Multi-level commands

You can map multiple command levels to a single executable name.

```bash
# pzero foo bar baz
# pzero will try pzero-foo-bar first
# if not found, it falls back to pzero-foo
# this usually means subcommands like "bar baz" are handled by pzero-foo itself
```

This allows you to organize team commands in a natural way, such as `pzero release publish` or `pzero company bootstrap`.

## Naming notes

Inside each command segment, `pzero` normalizes `-` to `_` before lookup.

For example:

* `pzero my-cmd` -> executable name `pzero-my_cmd`

To keep naming predictable, prefer simple command names or use `_` in the plugin executable when your command segment contains `-`.

## Recommended workflow

1. Build or place the plugin executable in a directory that is already in `PATH`
2. Follow the `pzero-<command>` naming rule
3. Add help output in the plugin itself, then use `pzero <command> --help` to view usage

Plugins are discovered dynamically, so they are not part of the built-in static command list printed by `pzero --help`.
