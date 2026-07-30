# temingo

[![Go Report Card](https://goreportcard.com/badge/thetillhoff/temingo)](https://goreportcard.com/report/thetillhoff/temingo)

This software aims to provide a simple but powerful templating mechanism.

The original idea was to create a simple static site generator, which is not as overloaded with "unnecessary functionality" as f.e. hugo.
The result, though, should not specifically be bound to website contents, as it can be used for any textfile-templating.

Temingo supports

- normal-type templates (== single-file-output templates) that will render to exactly one output file,
- partial-type templates (== partial templates) that can be included in other templates, also in other partials,
- meta-type templates (== multi-file-output templates) that can be used to render multiple output files,
- static files that will be copied to the output directory as is - respecting their location in the input directory filetree (except for `meta.yaml` files which are used for meta-type templates, and values files specified via `--valuesfile`),
- an ignore file (`.temingoignore`) that works similar to `.gitignore`, but for the templating process,
- a watch mechanism to trigger a rebuild of the output directory if necessary, which continuously checks if there are file changes in the input directory or the `.temingoignore`,
- an integrated webserver for local development,
- custom template values via CLI flags or YAML files,
- markdown content support with automatic HTML conversion,
- breadcrumb navigation support,
<!-- - HTML beautification for readable output. -->

## Installation

If you're feeling fancy:

```sh
curl -s https://raw.githubusercontent.com/thetillhoff/temingo/main/install.sh | sh
```

or manually from <https://github.com/thetillhoff/temingo/releases/latest>.

### Docker

You can use temingo as a Docker image from GitHub Container Registry:

```bash
# Pull the image
docker pull ghcr.io/thetillhoff/temingo

# Run temingo
docker run --rm -v "$(pwd):/workspace" -w /workspace ghcr.io/thetillhoff/temingo
```

**Multi-stage Dockerfile example:**

```dockerfile
# Build stage - render templates with temingo
FROM ghcr.io/thetillhoff/temingo AS builder
COPY src src
RUN temingo

# Final stage - serve the rendered site
FROM nginx:alpine
COPY --from=builder /workspace/output /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### GitHub Action

You can use temingo as a GitHub Action in your workflows. The action uses the Docker image internally for simplicity and cross-platform compatibility:

```yaml
- name: Render templates with temingo
  uses: thetillhoff/temingo
  with:
    inputDir: './src/' # Optional: location of template files (defaults to "./src/")
    outputDir: './output/' # Optional: where rendered files are written (defaults to "./output/")
    VALUES: | # Optional: key=value pairs for templates
      siteName=My Site
      author=John Doe
```

By default, the action reads from `./src/` and writes to `./output/`, cleaning up the output directory before rendering. You can customize the input and output directories as needed.

## Quick Start

```sh
# Initialize a new project
temingo init example

# Build templates from ./src to ./output
temingo

# Watch for changes and serve locally
temingo --watch --serve
```

## Core Concepts

### Templates

Temingo processes three types of template files:

#### Normal Templates

Normal templates (`*.template*`) are single-file-output templates that render to exactly one output file. The `.template` extension is removed from the output filename.

**Example:**

File: `src/index.template.html`

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Welcome</title>
  </head>
  <body>
    <h1>Welcome</h1>
    <p>Path: {{ .path }}</p>
  </body>
</html>
```

**Output:** `output/index.html` (the `.template` extension is removed)

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Welcome</title>
  </head>
  <body>
    <h1>Welcome</h1>
    <p>Path: index.html</p>
  </body>
</html>
```

#### Partial Templates

Partial templates (`*.partial*`) are reusable template snippets that can be included in other templates. Partials are automatically wrapped with `{{ define ... }}` blocks using their file path as the name. Include them using the `template` action.

**Example:**

File: `src/partials/header.partial.html`

```html
<header>
  <nav>
    <a href="/">Home</a>
    <a href="/about">About</a>
  </nav>
</header>
```

File: `src/index.template.html`

```html
<!DOCTYPE html>
<html>
  <body>
    {{ template "partials/header.partial.html" . }}
    <main>
      <h1>Content</h1>
    </main>
  </body>
</html>
```

**Output:** `output/index.html`

```html
<!DOCTYPE html>
<html>
  <body>
    <header>
      <nav>
        <a href="/">Home</a>
        <a href="/about">About</a>
      </nav>
    </header>
    <main>
      <h1>Content</h1>
    </main>
  </body>
</html>
```

The partial is automatically available as `"partials/header.partial.html"` and can be included in any template.

#### Metatemplates

Metatemplates (`*.metatemplate*`) are multi-file-output templates that generate multiple output files, one for each sibling subfolder containing a `meta.yaml` file.

**Example:**

File: `src/blog/index.metatemplate.html`

```html
<!DOCTYPE html>
<html>
  <head>
    <title>{{ .meta.name }} - Blog</title>
  </head>
  <body>
    <h1>{{ .meta.name }}</h1>
    <p>Content for {{ .meta.name }}</p>
  </body>
</html>
```

Directory structure:

```text
src/blog/
  index.metatemplate.html
  post1/
    meta.yaml  # name: "First Post"
  post2/
    meta.yaml  # name: "Second Post"
```

**Output:** This generates two files:

- `output/blog/post1/index.html`

```html
<!DOCTYPE html>
<html>
  <head>
    <title>First Post - Blog</title>
  </head>
  <body>
    <h1>First Post</h1>
    <p>Content for First Post</p>
  </body>
</html>
```

- `output/blog/post2/index.html`

```html
<!DOCTYPE html>
<html>
  <head>
    <title>Second Post - Blog</title>
  </head>
  <body>
    <h1>Second Post</h1>
    <p>Content for Second Post</p>
  </body>
</html>
```

### Static Files

All files that are not templates, partials, metatemplates, `meta.yaml` files, or values files (specified via `--valuesfile`) are copied to the output directory as-is, preserving their location in the directory structure.

### Metadata System

Temingo provides a hierarchical metadata system using `meta.yaml` files:

#### Metadata Hierarchy

Metadata is aggregated by iterating through folders from the input directory down to the folder containing the template file. Lower-level `meta.yaml` files are merged into parent ones, with child values overriding parent values.

#### Child Metadata

For each template file, temingo searches for all `meta.yaml` files in direct child subfolders (one level down) and makes them available as `.childMeta.<foldername>`. The key in the map is the folder name only (the last component of the path), not the full path. This enables dynamic navigation menus, listing pages, or iterating over child items.

**Important:** The key in the `childMeta` map is the folder name only (e.g., `post1`), not the full path. Since only direct children (one level down) are included, nested subfolders (e.g., `blog/posts/post1/meta.yaml`) are not included when processing `blog/index.template.html`. However, if you have multiple direct child folders with the same name (which would require different parent paths), the later one processed will overwrite the earlier one in the map. To avoid conflicts, ensure direct child folder names are unique.

**Example structure:**

```text
src/
  blog/
    index.template.html
    post1/
      meta.yaml  # { title: "First Post", date: "2024-01-01" }
    post2/
      meta.yaml  # { title: "Second Post", date: "2024-01-02" }
```

**In `blog/index.template.html`:**

```html
<ul>
  {{ range $folderName, $meta := .childMeta }}
  <li>
    <a href="/blog/{{ $folderName }}">{{ $meta.title }}</a>
    <span>{{ $meta.date }}</span>
  </li>
  {{ else }}
  <li>No posts yet</li>
  {{ end }}
</ul>
```

**Accessing individual children:**

```html
{{ if .childMeta.post1 }}
<p>Latest post: {{ .childMeta.post1.title }}</p>
{{ end }}
```

**Note:** This feature loads metadata from direct child directories only (one level down). Each child's metadata is merged with the parent metadata, so you can access both child-specific and inherited values.

### Markdown Content

If a template path (either as sibling or as child for metatemplates) contains a `content.md` file, it is automatically converted to HTML and made available as `.content` during the templating process.

### Breadcrumbs

Breadcrumbs represent the parent directory structure, excluding the directory containing the current `index.html` file. Each breadcrumb has:

- `Name`: The directory name
- `Path`: The full path to that directory (e.g., `/blog/` or `/blog/posts/`)

**Examples:**

- `index.html` → `[]` (empty)
- `blog/index.html` → `[]` (empty, no parent)
- `blog/posts/index.html` → `[{Name: "blog", Path: "/blog/"}]`
- `blog/posts/2024/index.html` → `[{Name: "blog", Path: "/blog/"}, {Name: "posts", Path: "/blog/posts/"}]`

**Template usage:**

```html
<nav aria-label="Breadcrumb">
  <a href="/">Home</a>
  {{ range .breadcrumbs }}
  <span>/</span>
  <a href="{{ .Path }}">{{ .Name }}</a>
  {{ end }}
</nav>
```

### Template Variables

The following variables are available in all templates:

```text
.path          -> string: path to template (within input directory)
.breadcrumbs   -> []Breadcrumb: breadcrumb objects with Name and Path fields
.meta          -> map[string]interface{}: aggregated metadata for current folder (merged from parent directories)
.childMeta     -> map[string]interface{}: metadata of direct child subfolders, key is the folder name
.<key>         -> string: custom values passed via --value flags or --valuesfile
.content       -> string: markdown content converted to HTML (if content.md exists)
```

## Template Functions

Temingo provides built-in template functions that can be used in your templates:

### `includeWithIndentation`

The `includeWithIndentation` function allows you to indent content by a specified number of spaces. This is particularly useful when including partials or other content that needs to match the indentation level of the surrounding context.

**Syntax:**

```go
{{ includeWithIndentation <amount_of_indentation_spaces> <content_to_indent> }}
```

**Parameters:**

- `indentation` (int): The number of spaces to indent each line
- `content` (string): The content to indent

**Example:**

```html
<div class="container">{{ includeWithIndentation 4 .content }}</div>
```

Or with a multi-line string variable:

```html
<pre>
{{ includeWithIndentation 2 .codeBlock }}
</pre>
```

This will indent each line of the content by the specified number of spaces, ensuring proper formatting in the output. This is particularly useful when you need to maintain indentation levels for code blocks, nested HTML structures, or when including content that should match the surrounding indentation.

### `concat`

The `concat` function concatenates multiple strings together into a single string. This is useful when you need to combine multiple string values or variables.

**Syntax:**

```go
{{ concat <string1> <string2> ... <stringN> }}
```

**Parameters:**

- `string1`, `string2`, ... `stringN` (string): One or more strings to concatenate

**Example:**

```html
<a href="{{ concat "https://example.com/" .path }}">Link</a>
```

Or with multiple variables:

```html
<div class="{{ concat "container " .theme " " .size }}">Content</div>
```

The function accepts any number of string arguments and concatenates them in order, returning a single combined string.

### `capitalize`

The `capitalize` function capitalizes the first letter of each word in a string.

**Syntax:**

```go
{{ capitalize <string> }}
```

**Parameters:**

- `string` (string): The string to capitalize

**Example:**

```html
{{ capitalize "hello world" }}
<!-- Output: "Hello World" -->

{{ capitalize .title }}
<!-- Output: "My Blog Post" if .title is "my blog post" -->
```

### `reverse`

The `reverse` function returns a slice in reverse order. This is useful when you need to iterate over a slice in the opposite direction or display items in reverse chronological order.

**Syntax:**

```go
{{ reverse <slice> }}
```

**Parameters:**

- `slice` ([]interface{}): The slice to reverse

**Example:**

```html
{{ range reverse .items }}
<div>{{ . }}</div>
{{ end }}
```

Or with a breadcrumb navigation in reverse order:

```html
{{ range reverse .breadcrumbs }}
<a href="{{ .Path }}">{{ .Name }}</a>
{{ end }}
```

The function returns a new slice with elements in reverse order. If the input is `nil`, it returns `nil`. If the input is an empty slice, it returns an empty slice.

### `filterBy`

Filters `.childMeta` by a field value. Entries where the field is absent are kept by default; only entries with a non-matching value are dropped.

**Syntax:** `{{ filterBy <field> <value> .childMeta }}`

**Example — hide WIP items with `publish: false` in their `meta.yaml`:**

```html
{{ range $index, $element := filterBy "publish" true .childMeta }}
```

Composable with `sortBy` and `reverse`:

```html
{{ range reverse (sortBy "date" (filterBy "publish" true .childMeta)) }}
```

### `sri`

Emits the integrity hash of a remote subresource, so the attribute does not have to be maintained by hand.

**Syntax:** `{{ sri <url> [<algorithm>] }}`

```html
<script src="https://cdn.example/lib/5.2.1/x.js"
        integrity="{{ sri "https://cdn.example/lib/5.2.1/x.js" }}"
        crossorigin="anonymous"></script>
```

The default algorithm is `sha384`. Pass another as a second argument - `sha256`, `sha384` and `sha512` are supported:

```html
{{ sri "https://cdn.example/x.js" "sha512" }}
```

`sri` accepts remote URLs only. A hash of a file temingo produced would protect nothing, because whoever can alter a same-origin file can alter the document carrying its hash.

The hash is fetched at build time, so a build using `sri` fails when the target is unreachable - there is no correct output without the hash. Note that this also means the hash is whatever the host served during that build, which protects your visitors against later tampering but not against a host already compromised at build time. A hash committed to the template is stronger; the missing-integrity check will tell you when one is absent.

<!-- ### Component template
- [ ] partials are included 1:1, components are automatically parsed as functions and args can be passed (see description below)
  - take all files in the `./src/components/*`, and create a map[string]interface{} aka map[filename-without-extension]interface{} // TODO is it the right type?
  - for each of those, register them as equally named functions that are then passed to the funcMap for templating
  - They can then be called with {{ filename-without-extension arg0 arg1 ... }} where the args have to be in the format of `key=value`.
  - The args will then be passed to the component template file (they cannot call partials, but partials can call them), where they are provided as a map[key]value.
  - if the filename points to a file in a subfolder, f.e. `{{ icon/github }}` those files are taken instead. -->

## Configuration & Options

### Ignore Files

Temingo respects ignored paths as described in `./.temingoignore`, which uses a similar syntax as `.gitignore`. The ignore file is automatically watched for changes when using `--watch`.

### Configuration File

Temingo supports a per-project configuration file (`.temingo.yaml` in the current working directory) to set default values for command-line flags. This allows you to configure project-specific settings without passing flags every time.

**Default location:** `./.temingo.yaml` (in the current working directory)

**Custom location:** Use the `--config` flag to specify a different path

**Configuration format:** YAML file with flag names as keys

**Example `.temingo.yaml`:**

```yaml
inputDir: 'src/'
outputDir: 'dist/'
templateExtension: '.tmpl'
verbose: true
value:
  - 'siteName=My Blog'
  - 'author=John Doe'
valuesfile:
  - 'base-values.yaml'
  - 'production-values.yaml'
```

**Priority order:**

1. Command-line flags (highest priority)
2. Configuration file values
3. Default values (lowest priority)

CLI flags always override configuration file values when both are provided. The configuration file is useful for setting project-specific defaults that can still be overridden on the command line.

### Custom Template Values

You can pass custom values to templates in two ways:

- **CLI flags**: `--value key=value` (can be specified multiple times)
- **YAML files**: `--valuesfile path/to/file.yaml` (can be specified multiple times)

Multiple values files are merged in order, with later files overriding earlier ones. CLI values always override values from files when both are provided. Values are accessible in templates via `.<key>`.

**Example:**

```sh
# Build with custom values
temingo --value siteName="My Blog" --value author="John Doe"

# Build with values from YAML file
temingo --valuesfile values.yaml

# Build with multiple values files (merged in order, later files override earlier ones)
temingo --valuesfile base-values.yaml --valuesfile production-values.yaml

# Build with values from file and override some via CLI
temingo --valuesfile values.yaml --value siteName="Override Name"
```

### Directory Validation

Temingo performs early validation of input and output directories before processing:

- Verifies that input directory exists and is a directory
- Verifies that output directory exists and is a directory (or creates it if it doesn't exist)
- If output directory is inside input directory, automatically adds it to the ignore list at runtime (for that single run) to prevent processing loops and prints a warning. The ignore file itself is not modified.
- If outputDir and inputDir are the same directory, it will check if --noDeleteOutputDir is set. If it is not set, it will return an error.

### Output Directory Management

The `--noDeleteOutputDir` flag preserves existing output directory contents instead of recreating it from scratch. This only overwrites the rendered template files, making it possible to have `inputDir==outputDir`.

### Beautify

HTML beautification is enabled by default. Automatically formats HTML output for better readability. Currently supports `.html` files.

### Integrated Webserver

The `--serve` / `-s` flag runs a simple integrated webserver that serves the output directory. The webserver listens only on `127.0.0.1` for security (local connections only) and can be combined with `--watch` for automatic rebuilds on file changes.

**Example:**

```sh
temingo --serve --watch
```

### Watch Mode

The `--watch` / `-w` flag enables automatic rebuilding when files change:

- Automatically rebuilds output when files change
- Watches input directory, `.temingoignore` file, and values files
- Can be combined with `--serve` for automatic rebuilds and local serving

### Project Initialization

The `temingo init` command generates sample projects:

- `temingo init example`: A basic example project with blog structure and components
- `temingo init test`: A comprehensive test project showcasing all temingo features including partials, metadata, markdown content, and metatemplates

Only creates files if the input directory doesn't already exist.

### Version Information

The `temingo version` command prints the current build version.

### Dry-run Mode

The `--dry-run` flag previews what would be built without actually writing files. Useful for testing and validation.

### Verbose Mode

The `--verbose` / `-v` flag enables detailed logging:

- Provides additional information about the rendering process
- Useful for debugging and understanding what temingo is doing

### Reference Checking

Every build reports references in the rendered output that are broken, unverifiable, or point at nothing the build produced:

- external URLs that respond with an error, redirect, or require authorisation
- external URLs that respond but send no `Access-Control-Allow-Origin`, which makes an `integrity` hash unverifiable and causes the browser to block the subresource
- cross-origin scripts and stylesheets with no `integrity` hash
- an `integrity` hash with no `crossorigin` attribute, which the browser blocks outright
- a cross-origin `@import`, which cannot be integrity-protected at all
- internal paths that no output file answers

An internal path resolves if the build writes that file, a directory holding an `index.html`, or the path with `.html` appended - temingo cannot know which form your server prefers, so any of them counts. Paths that only a server rewrite could satisfy are never reported: the check proves absence or stays silent.

URLs written as visible text - in a code sample, or inside an HTML comment - are never reported. Neither is a `form` action, which addresses a server route rather than a file.

Findings do not fail the build. Pass `--strict` (or set `strict: true`) to exit non-zero when any finding is reported, which is the intended CI configuration. Strict mode is fatal on unreachable and unresolvable hosts too, so a transient network fault fails the build and the remedy is to run it again.

Accept expected findings with an allow list:

```yaml
strict: true
allow:
  - url: https://paywalled.example/*      # accept every finding for these
  - url: https://redirecting.example/*
    checks: [redirect]                    # accept only the named categories
```

A trailing `*` covers everything under it, so `https://example.com/*` matches the whole host. A `*` in the middle of a pattern matches within one path segment, so `https://cdn.example/lib/*/x.js` pins the filename while accepting any version.

Categories are `status`, `gated`, `redirect`, `unreachable`, `missing-target`, `missing-integrity`, `missing-crossorigin`, `no-cors-header` and `unverified-import`. A URL whose entry names no categories is never requested at all.

A redirect is best fixed by replacing the reference with its target rather than allowlisting it.

External URLs are requested once each per process, so a watch session pays only on its first build.

## Usage Examples

### Basic Usage

```sh
temingo                                    # Build templates from ./src to ./output
temingo init example                       # Initialize with example project
temingo init test                          # Initialize with comprehensive test project
temingo version                            # Print current version
```

### Advanced Usage

```sh
# Build with custom directories and extensions
temingo --inputDir ./templates --outputDir ./dist --templateExtension .tmpl

# Use custom configuration file
temingo --config ./custom-config.yaml

# Watch for changes and serve locally
temingo --watch --serve

# Build without clearing output directory
temingo --noDeleteOutputDir

# Dry run to see what would be built
temingo --dry-run --verbose
```

### Command Line Options

```text
--config: Path to configuration file (default: ./.temingo.yaml in current directory).
--inputDir, -i, default "./src": Sets the path to the template-file-directory.
--outputDir, -o, default "./output": Sets the destination-path for the compiled templates.
--templateExtension, -t, default ".template": Sets the extension of the template files.
--metaTemplateExtension, -m, default ".metatemplate": Sets the extension of the metatemplate files. Automatically excluded from normally loaded templates.
--partialExtension, -c, default ".partial": Sets the extension of the partial files.
--metaFilename, default "meta.yaml": Sets the filename of the meta files.
--markdownFilename, default "content.md": Sets the filename for markdown content files.
--temingoignore, default ".temingoignore": Sets the path to the ignore file.
--value, multiple occurrences possible: Pass custom values to templates in key=value format.
--valuesfile, multiple occurrences possible: Path to a YAML file containing key-value pairs for the templates. Files are merged in order, with later files overriding earlier ones. `--value` flags take precedence over values from files.
--noDeleteOutputDir, default false: Don't delete the output directory before building.
--watch, -w, default false: Watches the inputDir and the temingoignore.
--serve, -s, default false: Serves the output directory with a simple webserver.
--dry-run, default false: If enabled, will not touch the outputDir.
--verbose, -v, default false: Enables the debug mode which prints more logs.
```

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and improvements.

## Development

### How to test

`go test ./...`

### Decisions / best practices

- Don't have global variables in a package -> they would be obstructed for the consumer and are not threadsafe
- Don't use functional options -> they require a lot of code / maintenance. Also, having functions to set a context object every time a function is called is tedious
- Use Context (called engine in this project). Not necessarily the go-context package, but implement "instance of package" as context and use that.
- For packages that have "global" variables / arguments, use Context (called "engine" in this project) as well.
