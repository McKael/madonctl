# Templates and Themes

This folder contains sample templates to customize madonctl output.

Feel free to contribute if you have nice templates or if you want to work on themes as well!

## Configuration

The template prefix directory can be set in the configuration file with the
'template_directory' setting, or with the `MADONCTL_TEMPLATE_DIRECTORY`
environment variable.\
If set, template files are looked up relatively from this repository first
(unless they are absolute paths or start with "./" or "../").

The themes are located in the `themes` directory, inside the base template
directory.
A theme is a collection of templates grouped in a theme directory (the name of
the directory is the name of the theme).\
E.g. `$template_directory/themes/ansi/`

## Templates

Here's an example using a template with ANSI color escape codes (for UNIX/Linux):

    madonctl timeline --limit 2 --template-file ansi-status.tmpl

## Themes

To use a theme, simply specify the theme name with the --theme flag (the
--output=theme flag is implied):

    madonctl timeline --limit=2 --theme=ansi
    madonctl accounts statuses --limit 5 --theme ansi

    madonctl --theme=ansi accounts notifications --list
    madonctl --theme=ansi stream

Currently, if a template is missing, madonctl will fall back to the _plain_
output format.  (In the future it might just fail with an error message.)
