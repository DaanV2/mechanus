# Websockets

yada yada (TODo)

## Paths

The path is `ws://<domain>:[port]/api/v1/screen` or `wss://<domain>:[port]/api/v1/screen` and for devices you can provide an id either via header, query parameter or path id:

device examples
```
ws://<domain>:[port]/api/v1/screen/[deviceid]
ws://<domain>:[port]/api/v1/screen?deviceid=<deviceid>
ws://<domain>:[port]/api/v1/screen with: `X-Device-ID`: `<deviceid>`
```

## Authentication

When connection, the JWT must be passed along for any player / admin.
An API key can be given for devices. And either `deviceid` query parameter or `X-Device-ID` must be given

## Headers

| Header            |                                                                                |
| ----------------- | ------------------------------------------------------------------------------ |
| `X-Connection-ID` | A UUID that has been generated for this specific connection, used for tracking |
| `X-Device-ID`     | A header used to tell what the device id is                                    |
