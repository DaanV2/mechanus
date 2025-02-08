# build client
FROM node AS client
WORKDIR /app

ADD ./client/package.json .
ADD ./client/package-lock.json .

RUN npm ci

COPY ./client .

RUN npm run build

# build server
FROM golang AS server
WORKDIR /app

COPY ./server/go.mod .
COPY ./server/go.sum .

RUN go mod download

COPY ./server .

RUN go build -o mechanus main.go

# Combine
FROM busybox AS final
#FROM gcr.io/distroless/static-debian12 AS final
WORKDIR /app
COPY --from=client /app/build /web
COPY --from=server /app/mechanus .

EXPOSE 8080
EXPOSE 8443
EXPOSE 8666

# Default config
ENV WEB_HOST=
ENV WEB_PORT=8080
ENV WEB_STATIC_FOLDER=/web
ENV LOG_FORMAT=json

ENTRYPOINT ["./mechanus", "server"]