FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN GOSUMDB=off go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/resume-generator .


FROM alpine:3.20

ENV TECTONIC_VERSION=0.16.9
ENV TECTONIC_CACHE_DIR=/var/cache/tectonic

RUN apk add --no-cache \
    ca-certificates \
    curl \
    tar \
    fontconfig \
    ttf-dejavu \
    && fc-cache -fv \
    && curl -fsSL "https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%40${TECTONIC_VERSION}/tectonic-${TECTONIC_VERSION}-x86_64-unknown-linux-musl.tar.gz" -o /tmp/tectonic.tar.gz \
    && tar -xzf /tmp/tectonic.tar.gz -C /usr/local/bin tectonic \
    && chmod +x /usr/local/bin/tectonic \
    && rm /tmp/tectonic.tar.gz \
    && mkdir -p "${TECTONIC_CACHE_DIR}" /tmp/tectonic-warmup

COPY --from=builder /out/resume-generator /usr/local/bin/resume-generator
WORKDIR /app
COPY templates ./templates
COPY resume.json ./

# Generate the real LaTeX document
RUN resume-generator \
    -input resume.json \
    -output /tmp/tectonic-warmup \
    -format latex

# Warm the Tectonic cache
RUN tectonic \
    --outdir /tmp/tectonic-warmup \
    /tmp/tectonic-warmup/resume.tex

# Verify everything is cached
RUN tectonic \
    --only-cached \
    --outdir /tmp/tectonic-warmup \
    /tmp/tectonic-warmup/resume.tex

# Clean up
RUN rm -rf /tmp/tectonic-warmup

EXPOSE 8080

CMD ["/usr/local/bin/resume-generator", "-server", ":8080"]