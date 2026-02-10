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
