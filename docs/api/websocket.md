# Websockets

yada yada (TODo)

## Paths

The path is `<ws | wss>://<domain>:[port]/api/v1/screen/{screenid}/{id}`

## Screen ID

A screen ID can be a UUID used for specific screen identifiers or one of the following `player`, `admin` or `viewer`

## ID

The user id or device id

## Authentication

When connection, the JWT must be passed along for any player / admin.
An API key can be given for devices. And either `deviceid` query parameter or `X-Device-ID` must be given

## Headers

| Header            |                                                                                |
| ----------------- | ------------------------------------------------------------------------------ |
| `X-Connection-ID` | A UUID that has been generated for this specific connection, used for tracking |
| `X-Device-ID`     | A header used to tell what the device id is                                    |
