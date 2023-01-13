# temingo

This software aims to provide a simple but powerful templating mechanism.

The original idea was to create a simple static site generator, which is not as overloaded with "unnecessary functionality" as f.e. hugo.
The result, though, should not specifically be bound to website stuff, as it can be used for any textfile-templating. -> At least when [#9](https://github.com/thetillhoff/temingo/issues/9) is resolved.

## Usage
```
temingo
temingo -w
temingo init -> will generate a sample project in the current folder
```

```
--valuesfile, -f, default: []string{"values.yaml"}, "Sets the path(s) to the values-file(s)."
--inputDir, -i, default "./": Sets the path to the template-file-directory."
--componentsDir, -p, default "components": Sets the path to the components-directory."
--outputDir, -o, default "output": Sets the destination-path for the compiled templates."
--staticDir, -s, default "static", -Sets the source-path for the static files."
--templateExtension, -t, default ".template": Sets the extension of the template files."
--singleTemplateExtension, default ".single.template": Sets the extension of the single-view template files. Automatically excluded from normally loaded templates."
--componentExtension, default ".component": Sets the extension of the component files." //TODO: not necessary, should be the same as templateExtension, since they are already distinguished by directory -> Might be useful when "modularization" will be implemented
--temingoignore, default ".temingoignore": Sets the path to the ignore file.
--watch, -w", default false: Watches the template-file-directory, components-directory and values-files.
--debug, -d", default false: Enables the debug mode.
```

temingo will by default:
- take the source files from folder `./src`
- write the rendered files into folder `./output/`
- take all `*.component` files as intermediate templates / snippets
  - their names must be unique. temingo will check this.
- take all `*.template` files to be rendered
  - their names must be unique. temingo will check this.
  - for each of those file, temingo will check their folder for any subfolders. If there are any, their names will be added to a list which is available in this "parent" template
    This means you can iterate over them and generate links for them.
    Check each folder if it contains a `meta.yaml` file. If yes, parse it and make it available in the "parent" template. (key=folder-name, value=`/*/meta.yaml` object)
- take all `*.metatemplate` files and use them for rendering in all of their subfolders that contain a `meta.yaml` file. Pass the object in that file to each metatemplate
- take all other files and copy them into the output folder as-is
- read configuration from a `~/.temingo.yaml` file and a `./.temingo.yaml` file
- pass an object to each template-rendering:
  ```
  breadcrumbs: [] -> path to location of template, split by '/'
  components: map[string]template -> all components
  metatemplate: map[string]object -> aggregated metadata of subfolders
  * (for example the `meta.yaml` object)
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
