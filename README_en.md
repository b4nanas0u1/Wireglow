# Wireglow

[Русский](README.md) | **English**

Wireglow is a desktop client for managing network connections powered by the [free-turn-proxy](https://github.com/hackdiaz-dev/free-turn-proxy) and [mihomo](https://github.com/MetaCubeX/mihomo) engines. It brings connection profiles, local proxy controls, and network session status into a graphical interface.

## Purpose

The project is intended for working with network infrastructure the user is authorized to access, as well as for developing and testing integrations between network components. Remote connection settings are provided by a profile imported by the user.

Use of the application must comply with applicable laws and the rules of the networks and services involved.

## Status

The project is under active development.

## Client features

- Import, store, and edit selected settings in free-turn-proxy profiles.
- Connect and disconnect sessions, and view their status and latency.
- Use a local HTTP/SOCKS proxy or switch to system proxy mode.
- Configure local addresses and ports for network components.
- Use the interface in Russian or English.

## Roadmap

- TUN mode for routing traffic without configuring a proxy in each application.
- Routing rules to send selected services directly and route all other traffic through the connection.
- Download free-turn-proxy releases from GitHub and choose which version to use.

## Debian package

Debian 13 amd64 builds include the UI, controller and both engines. Install a
release `.deb` with `sudo apt install ./wireglow_<version>_amd64.deb`.
Close Wireglow and wait for proxy recovery before upgrading or removing it.
User profiles and preferences survive upgrades and removal. See the
[build and installation guide](docs/en/development-setup.md#debian-13-amd64-package).

## Development

Instructions for installing tools, setting up the environment, and building locally on Windows and Linux are available [here](docs/en/development-setup.md).

## Acknowledgements

Thanks to the authors of:

- [free-turn-proxy](https://github.com/hackdiaz-dev/free-turn-proxy)
- [mihomo](https://github.com/MetaCubeX/mihomo)
