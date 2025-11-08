# temingo

This software aims to provide a simple but powerful templating mechanism.

The original idea was to create a simple static site generator, which is not as overloaded with "unnecessary functionality" as f.e. hugo.
The result, though, should not specifically be bound to website contents, as it can be used for any textfile-templating. -> At least when [#9](https://github.com/thetillhoff/temingo/issues/9) is resolved.

Temingo supports

- normal-type templates (== single-file-output templates) that will render to exactly one output file,
- partial-type templates (== partial templates) that can be included in other templates, also in other partials,
<!-- (- component-type templates (== component templates) that can be used for very often recurring elements like html buttons, where the css classes are set at one point, image embeddings, ...) -->
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
- `Path`: The full path to that directory (e.g., `/blog` or `/blog/posts`)

**Examples:**

- `index.html` → `[]` (empty)
- `blog/index.html` → `[]` (empty, no parent)
- `blog/posts/index.html` → `[{Name: "blog", Path: "/blog"}]`
- `blog/posts/2024/index.html` → `[{Name: "blog", Path: "/blog"}, {Name: "posts", Path: "/blog/posts"}]`

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

**Planned improvements:**

- [ ] Partial/conditional rerender for only changed files
- [ ] Skip unchanged static files
- [ ] Skip unchanged rendered files

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

# Watch for changes and serve locally
temingo --watch --serve

# Build without clearing output directory
temingo --noDeleteOutputDir

# Dry run to see what would be built
temingo --dry-run --verbose
```

### Command Line Options

<!-- Automate this snipped to be generated from the code at build time, or make otherwise sure this reflects the current state of the code -->

```text
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

## Future Optimizations

The following optimizations are planned or under consideration:

- File extension autodiscover
  - Add table in readme on which extensions are covered
  - Minimum are html, css and js. Nice would be svg and somehow image integration in webpages (webp conversion, auto replace in all src)

Temingo _can_ do (this should probably be put into a dedicated application ("website optimizer"?) which could also include submodules like minifyCss, minifyHtml, minifyJs, prettifyCss, prettifyHtml, prettifyJs):

- Content validation, for example check if the result is valid html according to the last file extension of the file. Supported extensions:
  - `.html`
  - `.css`
  - `.js`
- Content minification, for example for html files. Supported extensions:
  - `.html`
  - `.css`
  - `.js`
- Optimized media embedding, for example for images. Supported media:

  - images
  - svg (pregenerate different colors?)

- [ ] SHA256, SHA384, and SHA512 generation for files, for example `*.js` files, so they can be added to your csp config file, and nonces are supported.
- [ ] CSS beautification support
- [ ] JS beautification support
- [ ] Minify flag/setting
- [ ] Minify html, warn user if there are undefined css classes used
- [ ] Minify css, warn user if there are unused css classes
- [ ] Minify js
- [ ] Media & Media references optimization

## Ideas & Future Enhancements

This section contains ideas, TODOs, and potential future enhancements that are not yet implemented.

### Dynamic Metadata

- [ ] Pass global variables like datetime (globally equal renderTime only)

### Content Markdown

- [ ] Variables can be used in markdown, too. (Not sure if this makes sense yet)

### Watch Mode Improvements

- [ ] Partial/conditional rerender for only the changed files -> also only those changes will be printed in the logs
  - fileWatcher/Render should check if the renderedTemplate is actually different from the existing file (in output/) -> hash if the files exist, check rendered stuff only writeFile when an actual change occured -> take double care of files that are created newly / deleted
- [ ] Don't delete & copy when a static file hasn't changed. Maintain the necessary hashtable/s for static files in memory.
- [ ] If output folder isn't empty, generate hashlist during first build
- [ ] Don't delete & recreate rendered files when its contents haven't changed

### Configuration File

- [ ] Temingo by default reads configuration settings from a `~/.temingo.yaml` file and a `./.temingo.yaml` file.
- [ ] Verify config file

### HTML Parser & Validation

- [ ] HTML parser notes:
  - parent -> Node / node-ref
  - siblings -> []Node
  - attributes (contains, not equals) -> map[string]string
  - content -> string/[]Node
- [ ] PrettifyHtml, minifyHtml, and the Css and Js equivalents must be dedicated packages. If they need to be implemented manually, put them in dedicated repos.
- [ ] Fail on invalid folder names (special chars etc) -> might be better in verifyHtml()
- [ ] Templating engine should save a mapping of (inserted) line-numbers. That way, when the contents are verified (aka html/css/js is invalid for example), it can point to the correct file and line.

### Component Libraries

- [ ] Components can be packed into "component libraries", similar to a package.json. maybe `component.yaml`, `import.yaml` or `dependency.yaml`.
  - References are to git repos and tags therein.
  - Alternatively introduce a global registry for components, like godocs
  - Either helm-repo approach, or apt/godocs-approach
  - Local overrides should still be possible / components need to be able to be adjusted per project still
  - Maybe a `values.yaml` (optional) that can add additional properties/variables or overriding default ones for the whole lib
- [ ] Make it possible to print all css dependencies & overriding tree -> per component

### Minification & Optimization

- [ ] (div-merge on minifyHtml) // might clash with css rules...
- [ ] Automatically prettify generated files by default - or minify, depending on configuration

### Development Server

- [ ] Inform dev-server (serve? import as package?) via websocket, that there was a change. auto-include library for cache-reset and refresh websocket connection

### Other Ideas

- [ ] Use html meta tag for listview attributes

## TODO

- Test the rendering via golang tests, not manually.

- go through comments in README and todos in code

- move funcmap add to template engine into extra function, so it happens always exactly the same for the temporaryTemplateEngine and the templateEngine

- automatically check all "internal" links of website for validity aka file exists
- automatically check all links that have a protocol specified to use https and warn in case of http

- add setting to enable/disable auto-intendation of multiline partials with same whitespace as reference. Default is enabled.

## Development

### How to test

`go test ./...`

### Decisions / best practices

- Don't have global variables in a package -> they would be obstructed for the consumer and are not threadsafe
- Don't use functional options -> they require a lot of code / maintenance. Also, having functions to set a context object every time a function is called is tedious
- Use Context (called engine in this project). Not necessarily the go-context package, but implement "instance of package" as context and use that.
- For packages that have "global" variables / arguments, use Context (called "engine" in this project) as well.
