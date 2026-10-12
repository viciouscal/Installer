# Vencord Arab Installer

## Fork testing and releases

The GUI opens in Arabic. Press the borderless rounded flag beside US - AR to switch
languages during the session. Arabic display uses contextual letter shaping,
bidirectional text ordering and a bundled IBM Plex Sans Arabic Bold and IBM Plex Sans Bold fonts; Latin names, URLs and
paths retain their direction. CLI output and technical diagnostics remain in
English. IBM Plex's SIL Open Font License and the Unicode data license are in
`assets/`. Tests cover Arabic joining, mixed text, font glyph coverage and the
static GUI translations. Localized text is stored in logical order in
`assets/ar.json`; shaping is applied only for display.

This fork installs `viciouscal/Vencord` and checks installer updates from
`viciouscal/Installer`. If the GitHub REST API fails (including HTTP 403 or
429), release metadata is read from that same fork's public release page and
expanded asset list. This fallback depends on GitHub's HTML structure; it does
not bypass a block on github.com itself.

Pushes to `fix/**` branches and pull requests build downloadable Actions
artifacts only. Preview builds use version `Unknown`, disabling self updates
so the current stable release cannot replace a preview during testing.

Test in an isolated data directory using `VENCORD_USER_DATA_DIR`, then run
`go test -race -tags cli ./...`. Set `INSTALLER_LIVE_TEST=1` to also check both
forks' public release pages. Building or testing does not patch Discord; do
not select Install, Repair, or Uninstall until ready to test an installation.

After review, use a **new** version tag. The release workflow creates a draft
and refuses to overwrite an existing release for that tag. Existing releases,
tags, and their assets are retained. Publishing the draft is a separate manual
step. macOS signing uses configured Apple secrets when present; otherwise builds
use ad hoc signing. Notarization runs only when its credentials are configured. Winget submissions run only in the official repository.

The Vencord Arab Installer installs [Vencord Arab](https://github.com/viciouscal/Vencord)


## Usage

See https://github.com/viciouscal/Installer/releases

## Building from source

### Prerequisites 

You need to install the [Go programming language](https://go.dev/doc/install) and GCC, the GNU Compiler Collection (MinGW on Windows)

<details>
<summary>Additionally, if you're using Linux, you have to install some additional dependencies:</summary>

#### Base dependencies
```sh
apt install -y pkg-config libsdl2-dev libglx-dev libgl1-mesa-dev
dnf install pkg-config libGL-devel libXxf86vm-devel
```

#### X11 dependencies
```sh
apt install -y xorg-dev
dnf install libXcursor-devel libXi-devel libXinerama-devel libXrandr-devel
```

#### Wayland dependencies
```sh
apt install -y libwayland-dev libxkbcommon-dev wayland-protocols extra-cmake-modules
dnf install wayland-devel libxkbcommon-devel wayland-protocols-devel extra-cmake-modules
```

</details>

### Building

#### Install dependencies

```sh
go mod tidy
```

#### Build the GUI

##### Windows / Mac / Linux X11
```sh
make GUI=1
```

##### Linux Wayland
```sh
make GUI=1 WAYLAND=1
```

#### Build the CLI
```sh
make
```

You might want to pass some flags to this command to get a better build.
See [the GitHub workflow](https://github.com/viciouscal/Installer/blob/main/.github/workflows/release.yml) for what flags I pass or if you want more precise instructions
