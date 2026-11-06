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
FROM --platform=$BUILDPLATFORM golang:1.27.3 AS server
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app

COPY ./server/go.mod .
COPY ./server/go.sum .

RUN go mod download

COPY ./server .

# using mechanus-result because of collision with the folder
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o mechanus-result main.go
RUN go run ./scripts/licenses -o /app/licenses-go.json

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
