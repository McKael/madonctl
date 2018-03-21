# madonctl

Golang command line interface for the Mastodon API

[![license](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/McKael/madonctl/master/LICENSE)
[![Build Status](https://travis-ci.org/McKael/madonctl.svg?branch=master)](https://travis-ci.org/McKael/madonctl)

`madonctl` is a [Go](https://golang.org/) CLI tool to use the Mastondon REST API.

It is built on top of [madon](https://github.com/McKael/madon), my Golang
implementation of the API.

## Installation

### Download

Check the [Release page](https://github.com/McKael/madonctl/releases) for some
pre-built binaries.

More pre-built binaries might be available from the [Homepage](https://lilotux.net/~mikael/pub/madonctl/)
(development version and builds for a few other platforms).

### From source

To install the application from source (you need to have Go >= 1.7), just type:

    go get -u github.com/McKael/madonctl

and you should be able to run `madonctl`.

For upgrades, don't forget the `-u` option to ensure the dependencies
(especially [madon](https://github.com/McKael/madon)) are updated as well.

Travis automated testing is run for Go versions 1.7 - 1.10.

## Usage

### Configuration

In order to use madonctl, you need to specify the instance name or URL, and
usually provide an account login/password (or a token).

These settings can be passed as command line arguments or environment variables,
but the easiest way is to use a configuration file.

Note that you can **generate a configuration file** for your settings with

`madonctl config dump -i mastodon.social -L username@domain -P password`

(You can redirect the output to a configuration file.)

If you don't want to use the password or if you have enabled *Two-factor
authentication*, you can also use **OAuth2** with the `oauth2` command:

`madonctl -i mastodon.social oauth2 > madonctl.yaml`

For details on the configuration, please check the [configuration](configuration.md) page.

### Usage

The complete list of commands is available in the online help (`madonctl help`,
`madonctl command --help`...) or in the
[manpages](https://lilotux.net/~mikael/pub/madonctl/manual/html/).

### Examples

This is a non-exhaustive list of commands, showing how madonctl works.

To post a simple "toot":
``` sh
% madonctl toot "Hello, World"
```

You can change the toot visibility, add a Content Warning (a.k.a. spoiler)
or send a media file:
``` sh
% madonctl toot --visibility direct "@McKael Hello, you"
% madonctl toot --visibility private --spoiler CW "The answer was 42"
% madonctl post --file image.jpg Selfie # Send a media file
```
Note: The default toot visibility can be set in the configuration file with
the `default_visibility` setting or with the environment variable (example
`export MADONCTL_DEFAULT_VISIBILITY=unlisted`).

Send (text) file content as new message:
```
% madonctl toot --text-file message.txt
```

... or read message from standard input:
```
% echo "Hello from #madonctl" | madonctl toot --stdin
```

Reply to a message:
``` sh
% madonctl toot --in-reply-to 1234 --visibility direct "@user1 @user2 response"
% madonctl toot --in-reply-to 1234 --add-mentions "response"
```
The flag `--add-mentions` automatically adds mentions based on the toot you're
replying to.

Some **account-related commands**:
``` sh
% madonctl accounts blocked                       # List blocked accounts
% madonctl accounts muted                         # List muted accounts

% madonctl account notifications --list --all    # List really all notifications
% madonctl account notifications --list --clear  # List and clear notifications
% madonctl account notifications --notification-id 1234 # Display notification
% madonctl account notifications --dismiss --notification-id 1234
```

Note: By default, madonctl will send a single query.  If you want all available
results you should use the `--all` flag.  If you use a `--limit` value,
madonctl might send several queries until the number of results reaches this
value.

**Update** your account information:
``` sh
% madonctl account update --display-name "John"  # Update display name
% madonctl account update --note "Newcomer"      # Update user note (bio)
% madonctl account update --note ""              # Clear note
% madonctl account update --avatar me.png        # Update avatar
```

See your own **posts**:
``` sh
% madonctl account statuses                      # See last posts
% madonctl account statuses --all                # See all statuses
```

Display accounts you're **following** or your **followers**:
``` sh
% madonctl accounts following                     # See last following
% madonctl accounts following --all               # See all followed accounts
% madonctl accounts followers --limit 30          # Last 30 followers
```

Add/remove a **favourite**, **boost** a status...
``` sh
% madonctl status --status-id 416671 favourite    # Fave a status
% madonctl status --status-id 416671 boost        # Boost a status
```

**Pin/unpin** a status...
``` sh
% madonctl status --status-id 533769 pin          # Pin a status
% madonctl status --status-id 533769 pin --unset  # Unpin a status
```

Search for an account (only accounts known to your instance):
``` sh
% madonctl accounts search gargron
```

**Follow** an account with known ID:
``` sh
% madonctl account follow --account-id 1234
```

Follow a remote account:
``` sh
% madonctl account follow --remote Gargron@mastodon.social
```

**Search** for accounts, statuses or hashtags:
``` sh
% madonctl search gargron
% madonctl search mastodon
```

When the account ID is unknown, --user-id can be useful.\
You can specify the (instance-specific) account ID number (--account-id) or
the user ID (--user-id).  In the later case, madonctl will search for the
user so it must match exactly the ID known to your instance (without the
@domain suffix if the user is on the same instance).  The --user-id flag can
also contain an HTTP account URL.
``` sh
% madonctl account --user-id Gargron@mastodon.social -l5 statuses # Last 5 statuses
% madonctl account --user-id https://mastodon.social/@Gargron -l5 statuses # Same
```
Please note that the form "Gargron@mastodon.social" will only work if this
account is known by the instance.  If it is unknown, the https link should work
(and the Mastodon server will learn about the account).

Read **timelines**:
``` sh
% madonctl timeline                 # Display home timeline
% madonctl timeline public          # Display federated timeline
% madonctl timeline public --local  # Display public local timeline

% madonctl timeline --limit 3       # Display 3 latest home timeline messages
```

Use the **streaming API** and fetch timelines and notifications:
``` sh
% madonctl stream                   # Stream home timeline and notifications
% madonctl stream local             # Stream local timeline
% madonctl stream public            # Stream federated timeline
```

You can also use **hashtag streams**:
``` sh
% madonctl stream :mastodon         # Stream for hastag 'mastodon'
% madonctl stream :madonctl,golang  # Stream for several hashtags
```

Please note that madonctl will use one socket per stream, so the number of
concurrent hashtags is currently limited to 4 for "politeness".

It is also possible to send every stream event (notification or status) to
an **external command**.  You can can even combine it with a customized theme.
The contents will be sent to the standard input of this command:
``` sh
madonctl stream --command gateway.sh --theme gateway
```

(Almost) All commands have a **customizable output**:
``` sh
% madonctl account show            # Display an account
% madonctl account show -o yaml    # Display an account, in yaml
% madonctl account show -o json    # Display an account, in json
% madonctl stream local -o json     # Stream local timeline and output to JSON
```

You can also use Go (Golang) **templates**:
``` sh
% madonctl account --account-id 1 followers --template '{{.acct}}{{"\n"}}'
```

You can write and use [themes](templates) as well:
```
madonctl --theme=ansi timeline
```

There are many more commands, you can find them in the online help or the manpage.


### Shell completion

If you want **shell completion**, you can generate scripts with the following command: \
`madonctl completion bash` (or zsh)

Then, just source the script in your shell.

For example, I have this line in my .zshrc:

`source <(madonctl completion zsh)`

### Commands output

The output can be set to **json**, **yaml** or to a **Go template** for all commands.\
If you are familiar with Kubernetes' kubectl, it is very similar.

For example, you can display your user token with:\
`madonctl config whoami --template '{{.access_token}}'`\
or the application ID with:\
`madonctl config dump --template '{{.ID}}'`

All the users that have favorited a given status:\
`madonctl status --status-id 101194 favourited-by --template '{{.username}}{{"\n"}}'`

Sets of templates can be grouped as **themes**.

For more complex templates, one can use the `--template-file` option.\
See the [themes & templates](templates) folder.

## References

- [madonctl manpages](https://lilotux.net/~mikael/pub/madonctl/manual/html/)
- [madon](https://github.com/McKael/madon), the Go library for Mastodon API
- [Mastodon API documentation](https://github.com/tootsuite/documentation/blob/master/Using-the-API/API.md)
- [Mastodon repository](https://github.com/tootsuite/mastodon)
