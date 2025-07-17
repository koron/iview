# iview: Instant View local HTTP server

[![PkgGoDev](https://pkg.go.dev/badge/github.com/koron/iview)](https://pkg.go.dev/github.com/koron/iview)
[![Actions/Go](https://github.com/koron/iview/workflows/Go/badge.svg)](https://github.com/koron/iview/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/github.com/koron/iview)](https://goreportcard.com/report/github.com/koron/iview)

iview is an HTTP server that serves the contents of a local directory. When serving, it applies templates and filters depending on the file type. This makes it function as a pseudo file explorer or file viewer. Additionally, the file viewer has an auto-reload feature.

## Direcotry Structure for Resources

The resource direcotry (`_resource`) provides two `fs.FS`s: `static` and `template`

*   `static` directory provides static resources.
*   `template` direcotry provides templates.
    *   `layout.html` is the root layout template.
    *   others: templates for media types.

The resource directory is normally exposed via `embed.FS`, but the physical file system can be used for debugging purposes using the `-rsrc` option.

## Misc.

*   You can open a file in an editor.  The editor can be specified with the `-editor` flag, or the environment variables `IVIEW_EDITOR` and `EDITOR`.  The priority is as described above.

## Developer Resources

*   Inspect shared workers in Chrome

    `chrome://inspect/#workers`

    Copy the URL above. For security reasons, it cannot be opened as a link.

*   [**HTMX** Javascript API](https://htmx.org/api/)

*   [sindresorhus/github-markdown-css](https://github.com/sindresorhus/github-markdown-css)

*   [Material Symbols](https://fonts.google.com/icons?icon.set=Material+Symbols)

    *   Guide: <https://developers.google.com/fonts/docs/material_symbols>
    *   CSS base `_resource/static/thirdparty/material-symbols.css`:  
        <https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined>
    *   Font file `_resource/static/thirdparty/material-symbols.woff2`:  
        <https://fonts.gstatic.com/s/materialsymbolsoutlined/v257/kJF1BvYX7BgnkSrUwT8OhrdQw4oELdPIeeII9v6oDMzByHX9rA6RzaxHMPdY43zj-jCxv3fzvRNU22ZXGJpEpjC_1v-p_4MrImHCIJIZrDCvHOej.woff2>

*   [Syntax Highlighter: alecthomas/chroma](https://github.com/alecthomas/chroma)
