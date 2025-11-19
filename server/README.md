# Mechanus - Server

- [ ] Provides the files
- [ ] Login and users

## Development

You can either use the vscode launch configuration `launch server` or run the following commands:

```bash
make start-server
```

## Documentation

[Golang Docs](https://pkg.go.dev/github.com/DaanV2/mechanus/server)

### Data

The server uses SQLite for data and blob storage. because of the config it saves to `server/.local/db.sqlite`
