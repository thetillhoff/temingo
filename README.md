# temingo

This software aims to provide a simple but powerful templating mechanism.

The original idea was to create a simple static site generator, which is not as overloaded with "unnecessary functionality" as f.e. hugo.
The result, though, should not specifically be bound to website stuff, as it can be used for any textfile-templating. -> At least when [#9](https://github.com/tillhoff/temingo/issues/9) is resolved.


# notes for later docs
## help
- add a `--help` flag to get information about what options are available, what they are for and whether they have defaults.
## debug mode
- add a `--debug` flag to get information about what was done.
## single-view templates
- single-view templates are distinguished via their extension. Normal templates look like `*.ext.template` whereas single-view templates look like `*.ext.single.template`.
- single-view templates are templated in their dedicated step. So to prevent later problems, they are automatically excluded from the normal templating process.
