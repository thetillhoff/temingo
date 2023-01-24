# temingo

This software aims to provide a simple but powerful templating mechanism.

The original idea was to create a simple static site generator, which is not as overloaded with "unnecessary functionality" as f.e. hugo.
The result, though, should not specifically be bound to website contents, as it can be used for any textfile-templating. -> At least when [#9](https://github.com/thetillhoff/temingo/issues/9) is resolved.

Temingo supports
- normal-type templates (== single-output templates) that will render to exactly one output file,
- component-type templates (== partial templates) that can be included in other templates,
- meta-type templates (== multi-output templates) that will render to multiple output files,
- static files that will be copied to the output directory as is - respecting their location in the input directory filetree (except for `meta.yaml` files),
- an ignore file (`.temingoignore`) that works similar to `.gitignore`, but for the templating process.
- a watch mechanism, that continously checks if there are filechanges in the input directory or the `.temingoignore` - and trigger a rebuild if necessary.

## Usage
```
temingo
temingo init // Generates a sample project in the current folder. Only starts writing files if the input directory doesn't exist yet. Supports all flags except `--watch`.
```

```
<!-- --valuesfile, -f, default: []string{"values.yaml"}, "Sets the path(s) to the values-file(s)." // TODO adjust docs as its already implemented via meta.yaml -->
--inputDir, -i, default "./src": Sets the path to the template-file-directory."
--outputDir, -o, default "./output": Sets the destination-path for the compiled templates."
--templateExtension, -t, default ".template": Sets the extension of the template files."
--metaTemplateExtension, -m, default ".metatemplate": Sets the extension of the metatemplate files. Automatically excluded from normally loaded templates."
--componentExtension, -c, default ".component": Sets the extension of the component files."
--temingoignore, default ".temingoignore": Sets the path to the ignore file.
--watch, -w, default false: Watches the inputDir and the temingoignore.
--verbose, -v, default false: Enables the debug mode which prints more logs.
```

temingo will by default:
- take the source files from folder `./src`.
- consider the ignored paths as described in `./.temingoignore` which has a similar syntax as a `.gitignore`.
- write the rendered files into folder `./output`
- take all `*.component*` files as intermediate templates / snippets
  - the defined intermediate template names must be globally unique so they can be imported properly later. Temingo checks this.
- take all `*.template*` files to be rendered
  - for each of those file, temingo will check their folder for any subfolders. If there are any, their names will be added to a list which is available in this "parent" template
    This means you can iterate over them and generate links for them.
    Check each folder if it contains a `meta.yaml` file. If yes, parse it and make it available in the "parent" template. (key=folder-name, value=`/*/meta.yaml` object)
- take all `*.metatemplate*` files and use them for rendering in all of their subfolders that contain a `meta.yaml` file. Pass the object in that file to each metatemplate
- take all other files (static) and copy them into the output folder as-is. Except `meta.yaml`s.
- read configuration from a `~/.temingo.yaml` file and a `./.temingo.yaml` file
- metadata that is passed to the rendering will be aggregated as follows;
  - Iterate through folders from inputDir `./src` down to the folder containing the template file
  - On that way, always merge the lowerlevel `meta.yaml` (if it exists) into the parent one (overwrite if necessary)
  - Pass the final object to the respective template rendering process
- What else does this passed object contain that is passed to each template rendering process:
  ```
  ["path"] = string -> path to template (within `./src/`)
  ["breadcrumbs"] = []string -> path to location of template, split by '/'
  ["meta"] = map[string]object -> metadata for current folder
  ["childmeta"] = map[string]object -> aggregated metadata of subfolders, key is the folder name containing the respective metadata
  * (for example the `meta.yaml` object -> it'll start at the root folder, then start descending the folder-tree and grab all `meta.yaml`s along the way. Merged in a way that overwrites the parents' values.)
  ```

<!--
TODO
temingo _can_ do (alternatively this should be put into a dedicated application ("website optimizer"?) which could also include submodules like minifyCss, minifyHtml, minifyJs, prettifyCss, prettityHtml, prettifyJs):
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
-->

<!--
TODO
- write unit tests for temingo and fileIO
- Set proper cmd descriptions
- Write a comment/description for each method (temingo and fileIO)
- allow to override `meta.yaml` filename via cli flag
- Move fileIO into dedicated git-repo

- pass global variables like datetime (globally equal renderTime only)
- fileWatcher/Render should check if the renderedTemplate is actually different from the existing file (in output/) -> hash if the files exist, check rendered stuff only writeFile when an actual change occured -> take double care of files that are created newly / deleted
-->

<!--
html parser notes
- parent -> Node / node-ref
- siblings -> []Node
- attributes (contains, not equals) -> map[string]string
- content -> string/[]Node

- prettifyHtml, minifyHtml, and the Css and Js equivalents must be dedicated packages. If they need to be implemented manually, but them in dedicated repos.
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


# notes for later docs
## dry-run
- add a `--dry-run` flag to do all templating and so on, but dont interact with the outputDir at all.
