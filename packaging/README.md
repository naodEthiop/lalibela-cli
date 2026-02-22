# Package Manager Manifests

This folder contains manifests you can submit to Windows package managers.

## Winget

Manifest path:

`packaging/winget/manifests/NaodEthiop/Lalibela/0.1.23/`

Submit these files to:

`https://github.com/microsoft/winget-pkgs`

After approval, users can install with:

```powershell
winget install NaodEthiop.Lalibela
```

## Scoop

Manifest path:

`packaging/scoop/lalibela.json`

Users can install directly from this manifest URL:

```powershell
scoop install https://raw.githubusercontent.com/naodEthiop/lalibela-cli/main/packaging/scoop/lalibela.json
```
