### `gen-android` — Scaffold Android projects from a template

A single-binary CLI that generates a new Android project from an embedded template — no git clone, no network.

```
gen-android <package.name> <ProjectName> [--single]
```

**Examples:**

```bash
gen-android com.example.myapp MyApp          # multi-module template
gen-android com.example.myapp MyApp --single # single-module template
```

The `gen-android` binary embeds the Android project template at compile time via Go's `//go:embed`. It extracts the files into `./ProjectName/`, rewrites package names in all source files, and runs `git init`.

**Flags:**

| Flag       | Description                            |
| ---------- | -------------------------------------- |
| `--single` | Use the single-module template variant |

**Install:**

```bash
go install github.com/paulcoding810/android-template@latest
```

Or build from source:

```bash
git clone https://github.com/paulcoding810/android-template
cd android-template
go build -o gen-android .
```

### What's inside

The template includes a modern Android project with:

- Kotlin + Jetpack Compose
- Gradle Kotlin DSL
- Module structure: `app`, `core/*`, `feature/*` (multi-module) or flat `app/` (single-module)
- Pre-configured CI and proguard rules
