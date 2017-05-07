# Templates

This folder contains sample templates to customize madonctl output.

I'd like to eventually have support for themes with collections of templates.

Here's an example using a template with ANSI color escape codes (for UNIX/Linux):

    madonctl timeline --limit 2 --template-file ansi-status.tmpl

The template prefix directory can be set in the configuration file with the 'template_directory' setting,
or with the `MADONCTL_TEMPLATE_DIRECTORY` environment variable. \
If set, template files are looked up relatively from this repository first
(unless they are absolute paths or start with "./" or "../").

Feel free to contribute if you have nice templates or if you want to work on themes as well!
