## zipline add

Store a command under an alias

```
zipline add <alias> <command> [flags]
```

### Examples

```
  zipline add runlike "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro assaflavie/runlike"
```

### Options

```
  -d, --description string   description of the alias
      --force                overwrite an existing alias
  -h, --help                 help for add
```

### SEE ALSO

* [zipline](zipline.md)	 - Manage long docker run commands as shell aliases

