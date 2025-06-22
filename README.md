# Standard Backups

> [!WARNING]
> This repository and project is very early in development. It's currently in the prototyping stages and doesn't fully function. It's **not ready for production use at all**.

Standard backups is currently merely an idea trying to take shape. I'm actively working on building a first proof of concept and I'll see how things move out from there.

## The Idea

If you've ever wanted to setup backups on linux, you've probably quickly realized that there's a lot of great tools out there but as an admin you have to figure out two things:

1. When and how do you run those backup tools to produce correct backups.
1. How do you backup all the apps and services you run on your servers.

### Standard Orchestration and Pluggable Backup Tools

In practice, this leads you to build sketchy one-off scripts for your backups. The quality of those script will vary with your programming know-how and how much time and effort you put into them. This should be easier.

There are tools out there that handle this backup orchestration for you but they're bound to specific backup tools. That means that if you want to use `restic` for your backups you need to use one orchestration tool and if you want to use `borg` you need to use another. That's better but not ideal.

This ends up being a lot of programs that all pretty much do the same thing. Some are in-house private scripts while others are open source tools. The level of quality and maintenance across those can vary quite a bit depending on much time and effort maintainers have to give. This seems like a lot of duplicate work.

The idea behind `standard-backups` is to implement a standardized backup orchestration process once and then integrate backup tools in a pluggable way. We save on all the custom scripts and different orchestration tools. This addresses point 1 above.

### Pluggable and Pre-defined Backup Recipes

As a general rule of thumb, backing up an app or service usually involves a few simple steps:

1. Run some commands to prepare for the backup (optional)
1. Backup one or many directories
1. Run some commands to clean up after the backup (optional)

This varies from one app to the next. Some apps are nice enough to provide guides and instructions in their documentation while others leave you to reverse engineer things. In a classic scenario, it's the system administrator who's stuck with this task. Problem is that they might be not be well equipped to do that. After all, running and operating an app doesn't mean that you know enough inner plumbing to extract its data for backups.

Wouldn't it be nice if the people who know the most about the app were able to provide all of that for you?

With `standard-backups` the rules on how to backup a given app are defined in a separate manifest that can be provided by the app developer or maintainer. If that doesn't work, it can be provided by package maintainers or even the community. Separating things this way allows you to pull use recipes from trusted sources instead of having to figure out things on your own.

### Simple Backups for Admins

So what's left for system administrators? Simple high level configuration. Using `standard-backups`' config, admins can:

- choose which backup tools to use,
- define destinations which are different instances of those backup tools (ex: local vs remote),
- define sources which connect recipes with one or more targets.

From there, they run `standard-backups` on a schedule and leave it to do all the hard work.

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
