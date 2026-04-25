FROM golang:1.26.2-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS build
LABEL org.opencontainers.image.source="https://github.com/4ks-io/4ks"
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY ./apps/api ./apps/api
COPY ./libs/go ./libs/go
RUN CGO_ENABLED=0 go build -o 4ksapi ./apps/api/main.go

ARG VERSION=0.0.0
ENV VERSION=$VERSION
RUN echo "${VERSION}" > /VERSION

FROM scratch
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./libs/reserved-words/reserved-words ./reserved-words
COPY ./apps/api/casbin /apps/api/casbin
COPY --from=build /app/4ksapi /4ksapi
COPY --from=build /VERSION /VERSION
USER appuser
CMD ["/4ksapi"]
