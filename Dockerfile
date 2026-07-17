FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN GOSUMDB=off go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/resume-generator .

FROM alpine:3.20

ENV TECTONIC_VERSION=0.16.9

RUN apk add --no-cache ca-certificates curl tar \
    && curl -fsSL "https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%40${TECTONIC_VERSION}/tectonic-${TECTONIC_VERSION}-x86_64-unknown-linux-musl.tar.gz" -o /tmp/tectonic.tar.gz \
    && tar -xzf /tmp/tectonic.tar.gz -C /usr/local/bin tectonic \
    && chmod +x /usr/local/bin/tectonic \
    && rm /tmp/tectonic.tar.gz

WORKDIR /app

COPY --from=builder /out/resume-generator /usr/local/bin/resume-generator
COPY index.html ./
COPY resume.json ./
COPY templates ./templates

EXPOSE 8080

CMD ["/usr/local/bin/resume-generator", "-server", ":8080"]