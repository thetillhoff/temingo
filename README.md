# temingo

This software aims to provide a simple but powerful templating mechanism.

The original idea was to create a simple static site generator, which is not as overloaded with "unnecessary functionality" as f.e. hugo.
The result, though, should not specifically be bound to website contents, as it can be used for any textfile-templating. -> At least when [#9](https://github.com/thetillhoff/temingo/issues/9) is resolved.

Temingo supports
- normal-type templates (== single-output templates) that correlate to exactly one output file,
- component-type templates (== partial templates) that can be included in other templates,
- meta-type templates (== multi-output templates) that will render to multiple output files,
- static files that will be copied to the output directory as is - respecting their location in the input directory filetree,
- an ignore file (`.temingoignore`) that works similar to `.gitignore`, but for the templating process.
- a watch mechanism, that continously checks if there are filechanges in the input directory or the `.temingoignore` - and trigger a rebuild if necessary.

## Usage
```
temingo
temingo init // Generates a sample project in the current folder. Only starts writing files if the input directory doesn't exist yet. Supports all flags except `-w`.
```

```
--valuesfile, -f, default: []string{"values.yaml"}, "Sets the path(s) to the values-file(s)." // TODO -> doesn't even work yet
--inputDir, -i, default "./src": Sets the path to the template-file-directory."
--outputDir, -o, default "./output": Sets the destination-path for the compiled templates."
--templateExtension, -t, default ".template": Sets the extension of the template files."
--metaTemplateExtension, -m, default ".metatemplate": Sets the extension of the metatemplate files. Automatically excluded from normally loaded templates."
--componentExtension, -c, default ".component": Sets the extension of the component files."
--temingoignore, default ".temingoignore": Sets the path to the ignore file.
--watch, -w, default false: Watches the template-file-directory, components-directory and values-files. // TODO
--verbose, -v, default false: Enables the debug mode which prints more logs. // TODO
```

temingo will by default:
- take the source files from folder `./src`.
- consider the ignored paths as described in `./.temingoignore` which has a similar syntax as a `.gitignore`.
- write the rendered files into folder `./output`
- take all `*.component` files as intermediate templates / snippets
  - their names must be globally unique. temingo will check this.
- take all `*.template` files to be rendered
  - their names must be unique. temingo will check this.
  - for each of those file, temingo will check their folder for any subfolders. If there are any, their names will be added to a list which is available in this "parent" template
    This means you can iterate over them and generate links for them.
    Check each folder if it contains a `meta.yaml` file. If yes, parse it and make it available in the "parent" template. (key=folder-name, value=`/*/meta.yaml` object)
- take all `*.metatemplate` files and use them for rendering in all of their subfolders that contain a `meta.yaml` file. Pass the object in that file to each metatemplate
- take all other files (static) and copy them into the output folder as-is. Except `meta.yaml`s.
- read configuration from a `~/.temingo.yaml` file and a `./.temingo.yaml` file
- metadata that is passed to the rendering will be aggregated as follows;
  - Iterate through folders from inputdir `/` to the folder containing the template
  - On that way, always merge the lowerlevel `meta.yaml` (if it exists) into the parent one
  - Pass the final object as `values[meta]` to the respective template rendering process
- What else does the `values[string]object` map contain tha tis passed to each template rendering process:
  ```
  ["path"] = string -> path to template (within `./src/`)
  ["breadcrumbs"] = []string -> path to location of template, split by '/'
  ["meta"] = map[string]object -> metadata for current folder
  ["childmeta"] = map[string]object -> aggregated metadata of subfolders, key is the folder name containing the respective metadata
  * (for example the `meta.yaml` object -> it'll start at the root folder, then start descending the folder-tree and grab all `meta.yaml`s along the way. Merged in a way that overwrites the parents' values.)
  ```

<!--
TODO
temingo _can_ do (alternatively this should be put into a dedicated application ("website optimizer"?)):
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
Instead of doing os.Stat all the time, write functions that check against the existing internal filetree
-->

<!--
TODO
- add cli help for all commands
- write unit tests
- move WatchChanges (==whole filesystem watcher) to its own dedicated package, then pass the Render(...) call as an argument
-->

## Development
### Adding commands / subcommands
`cobra-cli add <command>`

### How to test
```
go test ./...
```


# notes for later docs
## help
- add a `--help` flag to get information about what options are available, what they are for and whether they have defaults.
## debug mode
- add a `--debug` flag to get information about what was done.
## single-view templates
- single-view templates are distinguished via their extension. Normal templates look like `*.ext.template` whereas single-view templates look like `*.ext.single.template`.
- single-view templates are templated in their dedicated step. So to prevent later problems, they are automatically excluded from the normal templating process.
