ARG GO_VERSION=1.21
FROM golang:${GO_VERSION}-alpine AS build

RUN apk add --no-cache git
WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./

ARG VERSION
ARG SHA
ARG DATE
RUN CGO_ENABLED=0 go build \
  -ldflags="\
  -X 'main.version=${VERSION}' \
  -X 'main.commit=${SHA}' \
  -X 'main.date=${DATE}' \
  -X 'main.builtBy=github'" \
  -o /smtp-pigeon ./cmd/smtp-pigeon
RUN /smtp-pigeon --version

FROM gcr.io/distroless/static AS final

ARG VERSION
ARG SHA
ARG DATE

LABEL org.opencontainers.image.created=${DATE}
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.revision=${SHA}
LABEL org.opencontainers.image.url="https://github.com/rktjmp/smtp-pigeon"
LABEL org.opencontainers.image.vendor="Oliver Marriott"
LABEL org.opencontainers.image.author="Oliver Marriott"
LABEL org.opencontainers.image.title="smtp-pigeon"
LABEL org.opencontainers.image.description="SMTP to HTTP POST with flexible configuration"

USER nonroot:nonroot
COPY --from=build --chown=nonroot:nonroot /smtp-pigeon /smtp-pigeon
ENTRYPOINT ["/smtp-pigeon"]
