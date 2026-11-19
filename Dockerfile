# Multi-arch image (linux/amd64 + linux/arm64). The client and server stages run on the build
# machine's platform and cross-compile; only the final stage is per target platform.

# build client
FROM --platform=$BUILDPLATFORM node:26 AS client
WORKDIR /app

ADD ./client/package.json .
ADD ./client/package-lock.json .

RUN npm ci --ignore-scripts

COPY ./client .

# also writes build/licenses/npm.json
RUN npm run build

# build server
FROM --platform=$BUILDPLATFORM golang:1.27.4 AS server
WORKDIR /app

COPY ./server/go.mod .
COPY ./server/go.sum .

RUN go mod download

COPY ./server .

# The licenses report is the same for every target platform, so it runs before TARGETOS /
# TARGETARCH are declared: BuildKit puts declared build args in the environment of every later
# RUN, which would make each platform rebuild this step (and the steps above) separately.
RUN --mount=type=cache,target=/root/.cache/go-build \
    go run ./scripts/licenses -o /app/licenses-go.json

# Only the compile itself is per target platform.
ARG TARGETOS
ARG TARGETARCH

# using mechanus-result because of collision with the folder
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o mechanus-result main.go
# Combine
FROM busybox AS final
# FROM gcr.io/distroless/static-debian12 AS final
WORKDIR /app
COPY --from=client /app/build /web
COPY --from=server /app/licenses-go.json /web/licenses/go.json
COPY --from=server /app/mechanus-result /app/mechanus


EXPOSE 8080
EXPOSE 8443

# Default config
ENV SERVER_HOST=
ENV SERVER_PORT=8080
ENV SERVER_STATIC_FOLDER=/web
ENV LOG_FORMAT=json

CMD ["/app/mechanus", "server"]
