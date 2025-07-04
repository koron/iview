# iview: Instant View local HTTP server

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/iview)](https://pkg.go.dev/github.com/koron/iview)
[![Actions/Go](https://github.com/koron/iview/workflows/Go/badge.svg)](https://github.com/koron/iview/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron/iview)](https://goreportcard.com/report/github.com/koron/iview)

iview is an HTTP server that serves the contents of a local directory. When serving, it applies templates and filters depending on the file type. This makes it function as a pseudo file explorer or file viewer. Additionally, the file viewer has an auto-reload feature.

## Direcotry Structure for Resources

The resource direcotry (`_resource`) provides two `fs.FS`s: `static` and `template`

* `static` directory provides static resources.
* `template` direcotry provides templates.
    * `layout.html` is the root layout template.
    * others: templates for media types.

The resource directory is normally exposed via `embed.FS`, but the physical file system can be used for debugging purposes using the `-rsrc` option.

## Developer Resources

*   Inspect shared workers in Chrome

    `chrome://inspect/#workers`

    Copy the URL above. For security reasons, it cannot be opened as a link.

*   [**HTMX** Javascript API](https://htmx.org/api/)

*   [sindresorhus/github-markdown-css](https://github.com/sindresorhus/github-markdown-css)
