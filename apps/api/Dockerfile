FROM golang:1.26.2-alpine AS build
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
