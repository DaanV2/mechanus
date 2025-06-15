# Mechanus - Server

- [ ] Provides the files
- [ ] Login and users

## Development

You can either use the vscode launch configuration `launch server` or run the following commands:

```bash
make start-server
```

### Data

The server uses SQLite for data and blob storage. because of the config it saves to `server/.local/db.sqlite`