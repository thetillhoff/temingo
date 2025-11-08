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

## Features

### Templating engine

Temingo by default:

- takes all `*.template*` files from the source folder `./src`.
- writes the rendered files into the destination folder `./output`.

### Ignoring source files

Temingo by default considers the ignored paths as described in `./.temingoignore` which has a similar syntax as a `.gitignore`.

### Support for static files / assets

Temingo by default takes all other files (static) and copies them into the output folder as-is. Except `meta.yaml`s and values files specified via `--valuesfile` (can be specified multiple times).

### Partial templates

Temingo by default takes all `*.partial*` files as intermediate templates / snippets

- [x] the defined intermediate template names must be globally unique so they can be imported properly later. Temingo verifies the uniqueness.
- [x] partials are added automatically with path, `partials/page.partial.html` is the automatic default name for that partial.
- [x] it's not needed to add the `{{define ...}} ... {{ end }}` part to partials, it's added automatically.
- [x] partials can be included using the built-in `template` action: `{{ template "partials/header.partial.html" . }}`
- [ ] allow globs for including templates, for example `{{ template "*.partial.css" . }}`, also for subfolders

### Template Functions

Temingo provides built-in template functions that can be used in your templates:

#### `includeWithIndentation`

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

#### `concat`

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

#### `capitalize`

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

### Dynamic metadata

Temingo by default passes the following metadata to the rendering:

- [ ] pass global variables like datetime (globally equal renderTime only)
- [x] `.path` contains the rendered template path
- [x] `.breadcrumbs` contains a slice of breadcrumb objects with `Name` and `Path` fields representing the folder hierarchy

#### Breadcrumbs

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

### Metadata hierarchy

Temingo by default aggregates the metadata that is passed to the rendering as follows;

- Iterate through folders from inputDir `./src` down to the folder containing the template file
- On that way, always merge the lowerlevel `meta.yaml` (if it exists) into the parent one (overwrite if necessary)
- Pass the final object to the respective template rendering process

### Metadata child list

For each `*.template*` file, temingo by default searches for all `./*/meta.yaml`s (in all folders that are one level further down from the template file) and adds them as `.childMeta.<foldername>.<content-object>` pair to the template.

This enables you to dynamically generate navigation menus, listing pages, or iterate over child items. Each child's metadata is automatically merged with the parent's metadata, so parent values are inherited.

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

### Metatemplates

Temingo by default takes all `*.metatemplate*` files and uses them as template in all of the sibling subfolders that contain a `meta.yaml` file. The object in those files are passed for each rendering.

### Content markdown

Temingo by default processes markdown files as follows:

- [x] If a template path (either as sibling or as child for the metatemplates) contains a `content.md` it is converted to html and made available as `.content` during the templating process.
<!-- - [ ] Variables can be used in markdown, too. (Not sure if this makes sense yet) -->

<!-- ### Configuration file -->
<!-- Temingo by default reads configuration settings from a `~/.temingo.yaml` file and a `./.temingo.yaml` file. -->
<!-- TODO verify config file support -->

### Watch mode

- [x] `--watch` / `-w` flag for watching for file changes in the source folder
  - [x] Automatically rebuilds output when files change
  - [x] Watches input directory, `.temingoignore` file, and values files
- [ ] partial/conditional rerender for only the changed files -> also only those changes will be printed in the logs
      fileWatcher/Render should check if the renderedTemplate is actually different from the existing file (in output/) -> hash if the files exist, check rendered stuff only writeFile when an actual change occured -> take double care of files that are created newly / deleted
- [ ] don't delete & copy when a static file hasn't changed. Maintain the necessary hashtable/s for static files in memory.
- [ ] if output folder isn't empty, generate hashlist during first build
- [ ] don't delete & recreate rendered files when its contents haven't changed

### Integrated simple webserver

- [x] `--serve` / `-s` flag for running a simple integrated webserver directly on the output folder
  - [x] Webserver only listens on `127.0.0.1` for security (local connections only)
  - [x] Can be combined with `--watch` for automatic rebuilds on file changes

### Project initialization

- [x] `temingo init example` command to generate sample project
- [x] Only creates files if the input directory doesn't already exist

### Version information

- [x] `temingo version` command to print the current build version

### Custom template values

- [x] `--value key=value` flag to pass custom values to templates
- [x] Multiple `--value` flags can be used to pass multiple key-value pairs
- [x] `--valuesfile` flag to load values from a YAML file (can be specified multiple times)
- [x] Multiple values files are merged in order, with later files overriding earlier ones
- [x] Values are accessible in templates via `.<key>`
- [x] CLI values override values from files when both are provided

### Dry-run mode

- [x] `--dry-run` flag to preview what would be built without actually writing files

### Verbose mode

- [x] `--verbose` / `-v` flag to enable detailed logging
- [x] Provides additional information about the rendering process
- [x] Useful for debugging and understanding what temingo is doing

### Output directory management

- [x] `--noDeleteOutputDir` flag to preserve existing output directory contents instead of recreating it from scratch.
      This only overwrites the rendered template files.
      Thus, it's possible to have inputDir==outputDir.

### Directory validation

- [x] Early validation of input and output directories before processing
- [x] Verifies that input directory exists and is a directory
- [x] Verifies that output directory exists and is a directory (or creates it if it doesn't exist)
- [x] Prevents output directory from being inside or equal to input directory

### Optimizations

- file extension autodiscover

  - add table in readme on which extensions are covered
  - minimum are html, css and js. nice would be are svg and somehow image integration in webpages (webp conversion, auto replace in all src)

  temingo _can_ do (this should probably be put into a dedicated application ("website optimizer"?) which could also include submodules like minifyCss, minifyHtml, minifyJs, prettifyCss, prettityHtml, prettifyJs):

  - content validation, for example check if the result is valid html according to the last file extension of the file. Supported extensions:
    - `.html`
    - `.css`
    - `.js`
  - content minification, for example for html files. Supported extensions:
    - `.html`
    - `.css`
    - `.js`
  - optimized media embedding, for example for images. Supported media:
    - images
    - svg (pregenerate different colors?)

- [ ] SHA256, SHA384, and SHA512 generation for files, for example `*.js` files, so they can be added to your csp config file, and nonces are supported.

#### Beautify

- [x] HTML beautification is enabled by default
- [x] Automatically formats HTML output for better readability
- [x] Supports `.html` files
- [ ] Supports `.css` files
- [ ] Supports `.js` files

#### Minify

TBD

- is it good to do this here? Wouldn't it be better to use something else instead? Linux approach, do one thing, but do it good.

- [ ] add flag / setting
- [ ] minify html, warn user if there are undefined css classes used
- [ ] minify css, warn user if there are unused css classes
- [ ] minify js

#### Media & Media references

TBD

- is it good to do this here? Wouldn't it be better to use something else instead? Linux approach, do one thing, but do it good.

- [ ] add flag / setting
- [ ] file extension autodiscover (html files only, which image format is used, depending on setting media format can be transformed as well)
- [ ] optimize media embedding automatically, but warn the user

## Usage

### Basic Usage

```sh
temingo                                    # Build templates from ./src to ./output
temingo init example                       # Initialize with example project
temingo init test                          # Initialize with comprehensive test project
temingo version                            # Print current version
```

### Advanced Usage

```sh
# Build with custom values
temingo --value siteName="My Blog" --value author="John Doe"

# Build with values from YAML file
temingo --valuesfile values.yaml

# Build with multiple values files (merged in order, later files override earlier ones)
temingo --valuesfile base-values.yaml --valuesfile production-values.yaml

# Build with values from file and override some via CLI
temingo --valuesfile values.yaml --value siteName="Override Name"

# Build with custom directories and extensions
temingo --inputDir ./templates --outputDir ./dist --templateExtension .tmpl

# Watch for changes and serve locally
temingo --watch --serve

# Build without clearing output directory
temingo --noDeleteOutputDir

# Dry run to see what would be built
temingo --dry-run --verbose
```

### Available Project Types

The `temingo init` command supports the following project types:

- `example`: A basic example project with blog structure and components
- `test`: A comprehensive test project showcasing all temingo features including partials, metadata, markdown content, and metatemplates

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

Here's a list of variables that are passed to each template rendering process:

```text
["path"] = string -> path to template (within `./src/`)
["breadcrumbs"] = []Breadcrumb -> breadcrumb objects with Name and Path fields (see Breadcrumbs section above for details)
["meta"] = map[string]object -> aggregated metadata for current folder (merged from parent directories)
["childMeta"] = map[string]object -> metadata of direct child subfolders, key is the folder name. Each child's metadata is merged with parent metadata. Use this to generate navigation menus or listing pages.
["<key>"] = map[string]string -> custom values passed via --value flags or --valuesfile
["content"] = string -> markdown content converted to HTML (if content.md exists)
```

## TODO

- Test the rendering via golang tests, not manually.

- go through comments in README and todos in code

- move funcmap add to template engine into extra function, so it happens always exactly the same for the temporaryTemplateEngine and the templateEngine

- automatically check all "internal" links of website for validity aka file exists
- automatically check all links that have a protocol specified to use https and warn in case of http

- add setting to enable/disable auto-intendation of multiline partials with same whitespace as reference. Default is enabled.

<!--
html parser notes
- parent -> Node / node-ref
- siblings -> []Node
- attributes (contains, not equals) -> map[string]string
- content -> string/[]Node

- prettifyHtml, minifyHtml, and the Css and Js equivalents must be dedicated packages. If they need to be implemented manually, put them in dedicated repos.
- fail on invalid folder names (special chars etc) -> might be better in verifyHtml()
- components can be packed into "component libraries", similar to a package.json. maybe `component.yaml`, `import.yaml` or `dependency.yaml`.
  - references are to git repos and tags therein.
  - alternatively introduce a global registry for components, like godocs
  - either helm-repo approach, or apt/godocs-approach
  - local overrides should still be possible / components need to be able to be adjusted per project still
  - maybe a `values.yaml` (optional) that can add additional properties/variables or overriding default ones for the whole lib
- make it possible to print all css dependencies & overriding tree -> per component
- use html <meta> tag for listview attributes
- (div-merge on minifyHtml) // might clash with css rules...
- templating engine should save a mapping of (inserted) line-numbers. That way, when the contents are verified (aka html/css/js is invalid for example), it can point to the corrent file and line.
- automatically prettify generated files by default - or minify, depending on configuration
- inform dev-server (serve? import as package?) via websocket, that there was a change. auto-include library for cache-reset and refresh websocket connection
-->

## Development

### Adding commands / subcommands

`cobra-cli add <command>`

### How to test

`go test ./...`

### Decisions / best practices

- Don't have global variables in a package -> they would be obstructed for the consumer and are not threadsafe
- Don't use functional options -> they require a lot of code / maintenance. Also, having functions to set a context object every time a function is called is tedious
- Use Context (called engine in this project). Not necessarily the go-context package, but implement "instance of package" as context and use that.
- For packages that have "global" variables / arguments, use Context (called "engine" in this project) as well.
