# Standard Backups

Standard Backups is a generic backup orchestration tool with a plugin-based
architecture that works with existing backup tools that you love and trust. It
handles all the boring logic (preparing backups, performing backups, cleanup,
secret management, etc.) and lets you focus on what you want to backup and where
you want those backups to go.

## Getting Started

Standard Backups uses backends to integrate with existing backup tools which
perform the backups. So you first need to choose which backend you'll be using.
Note that you can use more than one backend and you can change backends if you
change your mind.

The following backends are currently distributed alongside Standard Backups:

- [Restic](https://restic.net/): A fully featured backend that integrates with
  restic, a popular, fast, and secure backup program.
- Rsync: An example backend that uses rsync to backup files.

### Install

You'll need to install Standard Backups itself as well as a backend. Backends
are plugins that integrate with existing backup tools. You can use more than one
backend, but you'll need at least one.

1. Open the latest
   [release](https://github.com/dotboris/standard-backups/releases/latest) page.
1. Download the `standard-backups` package for your package manager and
   architecture.
1. Download at least one backend (`standard-backups-*-backend`) package for your
   package manager and architecture.
1. Install those packages with the instructions below (must be run as `root`).

<details>
<summary>DEB based Linux distributions (e.g. Ubuntu, Debian)</summary>

Go to the directory where you downloaded the `.deb` files for Standard Backups
and at least one backend. Then, run the following commands:

```sh
apt update
apt install ./standard-backups*.deb
```

</details>

<details>
<summary>RPM based Linux distributions (e.g. Fedora)</summary>

Go to the directory where you downloaded the `.rpm` files for Standard Backups
and at least one backend. Then, run the following commands:

```sh
dnf install --refresh ./standard-backups*.rpm
```

</details>

<details>
<summary>Alpine Linux</summary>

Go to the directory where you downloaded the `.apk` files for Standard Backups
and at least one backend. Then, run the following commands:

```sh
apk update
apk add --allow-untrusted ./standard-backups*.apk
```

</details>

<details>
<summary>Arch Linux</summary>

Go to the directory where you downloaded the `.pkg.tar.zst` files for Standard
Backups and at least one backend. Then, run the following commands:

```sh
pacman -Sy
pacman -U ./standard-backups*.pkg.tar.zst
```

</details>

To validate your installation, simply run `standard-backups list-backends`. This
should show you the backends that you have installed.

### Setup a Recipe

Recipes tell Standard Backups how to backup a given system, service, or
application. Each recipe consists of a list of paths to backup with exclusions,
an optional command to prepare the backup (before hook), an optional command to
cleanup the backup (after hook), and some metadata.

Standard Backups allows applications to distribute their own recipes. This saves
you from writing your own. You can see what recipes are available on your system
by running `standard-backups list-recipes`. If there's already a recipe for the
application you're trying to backup, take note of its name and move to the next
step. Otherwise, you'll need to write your own.

To make your own recipe, create a `.yaml` file under
`/etc/standard-backups/recipes/` with the following content:

```yaml
version: 1 # Internal, must be 1.
name: my-recipe # Name of your recipe. change this.
description: ... # Optional description of what your recipe does.
paths: # Paths that will be backed up. Change this.
  - /path/to/backup/...
  - /other/path/to/backup/...
exclude: # Optional paths that will not be backed up.
  - paths-not-to-backup
  - ...
before: # Optional command to run before the backup. Change or remove this.
  shell: bash # What shell to run the command through. (options: bash, sh)
  command: | # Commands to run. Change this.
    ... command to run ...
    ... supports multiple lines ...
after: # Optional command to run after the backup. Change or remove this.
  shell: bash # What shell to run the command through. (options: bash, sh)
  command: | # Commands to run. Change this.
    ... command to run ...
    ... supports multiple lines ...
```

Change this file to fit your needs following the comments. You can verify that
Standard Backups sees your recipe by running `standard-backups list-recipes`.

### Configure a Destination

Destinations are where backups go. Each is bound to a specific backend. As such,
configuration will change depending on what backend you choose. Follow the
example below that best fits your backend.

You can see which backends are installed by running `standard-backups list-backends`.

#### Restic Destination

Open `/etc/standard-backups/config.yaml` and add the following:

```yaml
destinations:
  my-destination: # Name of your destination. Change this.
    backend: restic
    options:
      repo: ... # Restic repo. Can be a local path or remote server / service.
      env:
        # Password for the restic repo. Don't put your password in clear-text here, use the secrets feature.
        RESTIC_PASSWORD: '{{ .Secrets.myDestinationPassword }}'
        # Add other environment variables needed by your repo here. Remember: don't put clear-text secrets here, use the secrets feature.
      forget: # Optional. Tells restic when to delete old backups.
        # Important: Read the restic guide on this feature before enabling it.
        # https://restic.readthedocs.io/en/stable/060_forget.html#removing-snapshots-according-to-a-policy
        enable: true
        options:
          keep-last: 4
          keep-daily: 7
          keep-weekly: 4
          keep-monthly: 12
          keep-yearly: 7

secrets:
  myDestinationPassword:
    # Where the repo password is stored.
    # - Create a file on your system and paste the password in there.
    # - Change its permissions so that only standard-backups can read it.
    # - Update the path here to match your file.
    from-file: /path/to/secret-file
```

#### Rsync Destination

> [!WARNING]
> The Rsync backend is not feature-rich or battle tested. It's not recommended for production use.

Open `/etc/standard-backups/config.yaml` and add the following:

```yaml
destinations:
  my-destination: # Name of your destination. Change this.
    backend: rsync
    options:
      destination-dir: ... # Where to store the backups.
```

### Perform Backups

First, you need to define a job. Jobs associate a recipe with one or more
destinations. To define a job, open `/etc/standard-backups/config.yaml` and add
the following:

```yaml
jobs:
  my-job: # Name of your job. Change this.
    recipe: ... # Name of the recipe you found or created earlier. Change this.
    backup-to: # Destinations where to send the backups.
      - my-destination # Destination we created earlier. Change this.
    on-success: # Optional command to run after the job succeeds. Change or remove this.
      shell: bash # What shell to run this command through. (options: bash, sh)
      command: | # Commands to run. Change this.
        ... command to run after job succeeds ...
    on-failure: # Optional command to run after the job fails. Change or remove this.
      shell: bash # What shell to run this command through. (options: bash, sh)
      command: | # Commands to run. Change this.
        ... command to run after job fails ...
```

You can now perform a backup by running `standard-backups backup my-job`. You
can see the resulting backup by running `standard-backups list-backups`.

Standard Backups doesn't provide a mechanism to run scheduled backups. Instead,
you are expected to use an existing task scheduling tool (`cron`, `systemd`
timers, etc.) to run `standard-backups backup ...` periodically.

It is recommended that you create a dedicated user for Standard Backups and
perform all backups as that one user. All files referenced in the `secrets`
section of the configuration should be owned and only readable by that user.

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
