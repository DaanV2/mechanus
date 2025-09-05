# build client
FROM node:24 AS client
WORKDIR /app

ADD ./client/package.json .
ADD ./client/package-lock.json .

ENV PLAYWRIGHT_SKIP_BROWSER_DOWNLOAD=1
RUN npm ci

COPY ./client .

RUN npm run build

# build server
FROM golang:1.25.1 AS server
WORKDIR /app

COPY ./server/go.mod .
COPY ./server/go.sum .

RUN go mod download

COPY ./server .

# using mechanus-result because of collision with the folder
RUN go build -o mechanus-result main.go

# Combine
FROM busybox AS final
# FROM gcr.io/distroless/static-debian12 AS final
WORKDIR /app
COPY --from=client /app/build /web
COPY --from=server /app/mechanus-result /app/mechanus


EXPOSE 8080
EXPOSE 8443
EXPOSE 8666

# Default config
ENV WEB_HOST=
ENV WEB_PORT=8080
ENV WEB_STATIC_FOLDER=/web
ENV LOG_FORMAT=json

CMD ["/app/mechanus", "server"]