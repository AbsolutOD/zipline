## zipline run

Run a stored alias, appending any extra arguments

### Synopsis

Run the command stored under <alias>, appending any extra arguments.
The zipline process is replaced by the command (exec), so interactive
containers, signals, and exit codes behave as if you ran it directly.

```
zipline run <alias> [-- <args>...] [flags]
```

### Options

```
  -h, --help   help for run
```

### SEE ALSO

* [zipline](zipline.md)	 - Manage long docker run commands as shell aliases

