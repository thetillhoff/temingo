# temingo

This software aims to provide a simple but powerful templating mechanism.

The original idea was to create a simple static site generator, which is not as overloaded with "unnecessary functionality" as f.e. hugo.
The result, though, should not specifically be bound to website contents, as it can be used for any textfile-templating. -> At least when [#9](https://github.com/thetillhoff/temingo/issues/9) is resolved.

Temingo supports

- normal-type templates (== single-file-output templates) that will render to exactly one output file,
- partial-type templates (== partial templates) that can be included in other templates, also in other partials,
<!-- (- component-type templates (== component templates) that can be used for very often recurring elements like html buttons, where the css classes are set at one point, image embeddings, ...) -->
- meta-type templates (== multi-file-output templates) that can be used to render multiple output files,
- static files that will be copied to the output directory as is - respecting their location in the input directory filetree (except for `meta.yaml` files which are used for meta-type templates),
- an ignore file (`.temingoignore`) that works similar to `.gitignore`, but for the templating process.
- a watch mechanism to trigger a rebuild of the output directory if necessary, which continously checks if there are filechanges in the input directory or the `.temingoignore`/

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

Temingo by default takes all other files (static) and copies them into the output folder as-is. Except `meta.yaml`s.

### Partial templates

Temingo by default takes all `*.partial*` files as intermediate templates / snippets

- [x] the defined intermediate template names must be globally unique so they can be imported properly later. Temingo verifies the uniqueness.
- [x] partials are added automatically with path, `partial/page.partial.html` is the automatic default name for that partial.
- [x] it's not needed to add the `{{define ...}} ... {{ end }}` part to partials, it's added automatically.
- [ ] allow globs for including templates, for example `{{ template "*.partial.css" . }}`, also for subfolders

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
- [x] `.meta.path` contains the rendered template path
- [x] `.meta.breadcrumbs` contains a slice of the folder hierarchy for each template

### Metadata hierarchy

Temingo by default aggregates the metadata that is passed to the rendering as follows;

- Iterate through folders from inputDir `./src` down to the folder containing the template file
- On that way, always merge the lowerlevel `meta.yaml` (if it exists) into the parent one (overwrite if necessary)
- Pass the final object to the respective template rendering process

### Metadata child list

For each `*.template*` file, temingo by default searches for all `./*/meta.yaml`s (in all folders that are one level further down from the template file) and adds them as `.childMeta.<foldername>.<content-object>` pair to the template.
This means you can iterate over them and for example generate links for them.

<!-- optional TODO have a path that can be set in the template, for which the files can be read -->

### Metatemplates

Temingo by default takes all `*.metatemplate*` files and uses them as template in all of the sibling subfolders that contain a `meta.yaml` file. The object in those files are passed for each rendering.

### Content markdown

Temingo by default processes markdown files as follows:

- [x] If a template path (either as sibling or as child for the metatemplates) contains a `content.md` it is converted to html and made available as `.content` during the templating process.
<!-- - [ ] Variables can be used in markdown, too. (Not sure if this makes sense yet) -->

<!-- ### Configuration file -->
<!-- Temingo by default reads configuration settings from a `~/.temingo.yaml` file and a `./.temingo.yaml` file. -->
<!-- TODO verify config file support -->

### Watch-mode

- [x] add --watch / -w flag for watching for file changes in the source folder
- [ ] partial/conditional rerender for only the changed files -> also only those changes will be printed in the logs
      fileWatcher/Render should check if the renderedTemplate is actually different from the existing file (in output/) -> hash if the files exist, check rendered stuff only writeFile when an actual change occured -> take double care of files that are created newly / deleted
- [ ] don't delete & copy when a static file hasn't changed. Maintain the necessary hashtable/s for static files in memory.
- [ ] if output folder isn't empty, generate hashlist during first build
- [ ] don't delete & recreate rendered files when its contents haven't changed

### Integrated simple webserver

- [x] add --serve / -s flag for running a simple integrated webserver directly on the output folder.

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

TBD

- is it good to do this there? Wouldn't it be better to use something else instead? Linux approach, do one thing, but do it good.

This is currently enabled by default.

- [ ] add flag / setting
- [x] beautify html
- [ ] beautify css
- [ ] beautify js

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

```sh
temingo
temingo init // Generates a sample project in the current folder. Only starts writing files if the input directory doesn't exist yet. Supports all flags except `--watch`.
```

<!-- Automate this snipped to be generated from the code at build time, or make otherwise sure this reflects the current state of the code -->
```
--inputDir, -i, default "./src": Sets the path to the template-file-directory.
--outputDir, -o, default "./output": Sets the destination-path for the compiled templates.
--templateExtension, -t, default ".template": Sets the extension of the template files.
--metaTemplateExtension, -m, default ".metatemplate": Sets the extension of the metatemplate files. Automatically excluded from normally loaded templates.
--partialExtension, -c, default ".partial": Sets the extension of the partial files.
--metaFilename, default "meta.yaml": Sets the filename of the meta files.
--temingoignore, default ".temingoignore": Sets the path to the ignore file.
--watch, -w, default false: Watches the inputDir and the temingoignore.
--dry-run, default false: If enabled, will not touch the outputDir.
--verbose, -v, default false: Enables the debug mode which prints more logs.
```

Here's a list of variables that are passed to each template rendering process:

```
["path"] = string -> path to template (within `./src/`)
["breadcrumbs"] = []string -> path to location of template, split by '/'
["meta"] = map[string]object -> aggregated metadata for current folder
["childMeta"] = map[string]object -> metadata of subfolders, key is the folder name containing the respective metadata
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
