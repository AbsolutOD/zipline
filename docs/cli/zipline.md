## zipline

Manage long docker run commands as shell aliases

### Synopsis

Zipline stores long commands (docker run, podman run, ...) under short
aliases, tracks how often you use them, and hooks into bash/zsh so each
alias works like a native command.

Add the hook to your shell rc file:

  eval "$(zipline init zsh)"    # or: zipline init bash

Then:

  zl add runlike "docker run --rm -v /var/run/docker.sock:/var/run/docker.sock:ro assaflavie/runlike"
  runlike <container>

### Options

```
  -h, --help   help for zipline
```

### SEE ALSO

* [zipline add](zipline_add.md)	 - Store a command under an alias
* [zipline edit](zipline_edit.md)	 - Edit an alias's command in $EDITOR
* [zipline init](zipline_init.md)	 - Print the shell hook script (add eval "$(zipline init zsh)" to your rc file)
* [zipline list](zipline_list.md)	 - List stored aliases
* [zipline remove](zipline_remove.md)	 - Remove an alias
* [zipline run](zipline_run.md)	 - Run a stored alias, appending any extra arguments
* [zipline show](zipline_show.md)	 - Show full details of an alias

