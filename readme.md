# JSExt Initialization Tool

Usage:
```
  -init
        Initialize a project
  -n string
        Name of the project to initialize.
  -plain
        Create a plain project
  -vscode
        Create a vscode config file.
```

## Example

```
$ jsext -init -n myproject -vscode
$ jsext -plain -n myproject -vscode (Does not generate example application.)
```

## Installation

```
$ go install github.com/Nigel2392/jsexttool
```