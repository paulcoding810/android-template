### `gen-android` — Scaffold Android projects from a template

A single-binary CLI that generates a new Android or Kotlin Multiplatform project from an embedded template — no git clone, no network.

```bash
gen-android <package.name> <ProjectName> [flags]
```

**Examples:**

```bash
gen-android com.example.myapp MyApp                          # multi-module template
gen-android com.example.myapp MyApp --single                 # single-module template
gen-android com.example.myapp MyApp --kmm --team-id ABCDE12345
gen-android com.example.myapp MyApp --output ~/projects --no-git
gen-android --list
```

The `gen-android` binary embeds the project templates at compile time via Go's `//go:embed`. It extracts the files into `<output>/ProjectName/`, rewrites package and project names in all paths and source files, and runs `git init` with an initial commit.

**Flags:**

| Flag               | Description                                              |
| ------------------ | -------------------------------------------------------- |
| `--single`         | Use the single-module template                           |
| `--kmm`            | Use the Kotlin Multiplatform template (Android, iOS, JVM) |
| `--output <dir>`   | Parent directory for the new project (default `.`)       |
| `--team-id <id>`   | Apple team ID written to the iOS config (`--kmm` only)   |
| `--no-git`         | Skip `git init` and the initial commit                   |
| `--list`           | List available templates                                 |

Flags may appear before or after the positional arguments. The package name must have at least two lowercase segments and no Kotlin keywords; the project name must be letters and digits only.

**Install:**

```bash
go install github.com/paulcoding810/android-template@latest
```

Or build from source:

```bash
git clone https://github.com/paulcoding810/android-template
cd android-template
make build
```

`make build` deletes local build output (`build/`, `.gradle/`, `.idea/`, `local.properties`, …) from the template folders first, because `go:embed` cannot exclude paths and would otherwise bloat the binary. The generator also skips these paths when writing a project.

**Develop:**

```bash
make test   # generates every template into a temp dir and checks renaming
```

**Keep templates current:**

```bash
make upstream-diff                          # compare template-single with `android create`
make upstream-diff TEMPLATE=template-main
```

Generates a fresh project with the [Android CLI](https://developer.android.com/tools), lists library and plugin versions that differ (matched by Maven coordinate), then diffs the build config files. Requires the `android` CLI and Python 3.11+. The KMM template should be compared against the Kotlin Multiplatform wizard instead.

### What's inside

| Template          | Flag        | Contents                                              |
| ----------------- | ----------- | ----------------------------------------------------- |
| `template-main`   | (default)   | Multi-module Android app: `app`, `core/*`, `feature/*` |
| `template-single` | `--single`  | Single-module Android app: `app/`                     |
| `template-kmm`    | `--kmm`     | Kotlin Multiplatform: `shared`, `androidApp`, `iosApp`, `jvmApp` |

All templates use Kotlin, Jetpack Compose (Compose Multiplatform for KMM), Gradle Kotlin DSL and proguard rules.
