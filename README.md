# Standard Backups

Standard Backups is a generic backup orchestration tool with a plugin
architecture that lets it work with existing backup tools that you love and
trust. It handles all the boring logic (preparing backups, performing backups,
cleanup, secret management, etc.) and lets you focus on what you want to backup
and where you want those backups to go.

## Getting Started

Standard Backups uses backends to integrate with existing backup tools which
perform the backups. So you first need to choose which backend you'll be using.
Note that you can use more than one backend and you can change backends if you
change your mind.

The following backends are currently distributed along side Standard Backups:

- [Restic](https://restic.net/): A fully featured backend that integrates with
  restic, a popular, fast, and secure backup program.
- Rsync: An example backend that uses rsync to backup files.

### Install

You'll need to install `standard-backups` itself as well as a backend. Backends
are plugins that integrate with existing backup tools. You can use more than one
backend, but you'll need at least one.

1. Open the latest
   [release](https://github.com/dotboris/standard-backups/releases/latest) page.
1. Download the `standard-backups` package for your package manager and
   architecture.
1. Download at least one backend (`standard-backups-*-backend`) package for your
   package manager and architecture.
1. Install those packages with the instructions below. Must be run as `root`.

<details>
<summary>DEB based Linux distributions (e.g. Ubuntu, Debian)</summary>

Go to the directory where you downloaded the `.deb` files for Standard Backups
and at least one backend. Then, run the following commands:

```sh
apt update
apt install ./standard-backups-*.deb
```

</details>

<details>
<summary>RPM based Linux distributions (e.g. Fedora)</summary>

Go to the directory where you downloaded the `.rpm` files for Standard Backups
and at least one backend. Then, run the following commands:

```sh
dnf install --refresh ./standard-backups-*.rpm
```

</details>

<details>
<summary>Alpine Linux</summary>

Go to the directory where you downloaded the `.apk` files for Standard Backups
and at least one backend. Then, run the following commands:

```sh
apk update
apk add --allow-untrusted ./standard-backups-*.apk
```

</details>

<details>
<summary>Arch Linux</summary>

Go to the directory where you downloaded the `.pkg.tar.zst` files for Standard
Backups and at least one backend. Then, run the following commands:

```sh
pacman -Sy
pacman -U ./standard-backups-*.pkg.tar.zst
```

</details>

To validate your installation, simply run `standard-backups list-backends`. This
should show you the backends that you have installed.

### Setup a Recipe

Recipes tell Standard Backups how to backup a given system, service, or
application. Each recipe consists of a list of paths to backup with exclusions,
an optional command to prepare the backup (before hook), an optional command to
cleanup the backup (after hook), and some metadata.

Standard Backups allows packages to ship their own recipes. This saves your from
writing your own. You can see what recipes are available on your system by
running `standard-backups list-recipes`. If there's already a recipe for the
application you're trying to backup take note of its name and move to the next
step. Otherwise, you'll need to write your own.

To make your own recipe, create a `.yaml` file under
`/etc/standard-backups/recipes/` with the following content:

```yaml
version: 1 # internal; must be 1
name: my-recipe # name of your recipe; change this
description: ... # optional description of what your recipe does
paths: # paths that will be backed up; change this
  - /path/to/backup/...
  - /other/path/to/backup/...
exclude: # optional paths that will not be backed up
  - paths-not-to-backup
  - ...
before: # optional command to run before the backup; change or remove this
  shell: bash # what shell to run the command through; (options: bash, sh)
  command: | # commands to run; change this
    ... command to run ...
    ... supports multiple lines ...
after: # optional command to run after the backup; change or remove this
  shell: bash # what shell to run the command through; (options: bash, sh)
  command: | # commands to run; change this
    ... command to run ...
    ... supports multiple lines ...
```

Change this file to fit your needs following the comments. You can verify that
Standard Backups sees your recipe by running `standard-backups list-recipes`.

### Configure a Destination

TODO: Restic
TODO: ???

### Perform Backups

TODO: Run a backup once
TODO: setup recurring backups

## License

Copyright (C) 2025 Boris Bera

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
