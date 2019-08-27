FROM golang:1.12.0-stretch as gobuilder
RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates
RUN go get -u golang.org/x/lint/golint

RUN addgroup --gid 990 app && adduser --disabled-password --uid 991 --gid 990 --gecos '' app

RUN mkdir -p /build
WORKDIR /build

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 go build -i -v -o release/fm -ldflags="-X main.version=0.1" cmd/*.go


FROM scratch
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuilder /etc/passwd /etc/passwd
COPY --chown=990:990 --from=gobuilder /build/release/fm /app

USER 990:990
ENTRYPOINT ["./app"]
CMD ["run"]
