# Integration tests

## Setup

Run `make setup` to step the tests and playwright. Ensure you do `make image` in the root to ensure the server image is made

## Testing

`make test` should be fine

## Developing

First ensure the docker image in the root folder is built.
Then before you start the tests, run `make server` to start up the server.
If no image exists then run `make image` first
Then you can happy develop integration tests.