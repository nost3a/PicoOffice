# PicoOffice multi-stage build
# Build context = repo root; this file lives at repo root.
#
#  builder: golang image compiles backend (CGO sqlite needs gcc)
#  runtime: debian + libreoffice + CJK fonts, runs soffice export

# ---------- builder ----------
FROM golang:1.26-bookworm AS builder

WORKDIR /src

# CGO_ENABLED=1 sqlite build needs gcc
RUN apt-get update && apt-get install -y --no-install-recommends gcc && rm -rf /var/lib/apt/lists/*

# copy dependency lists first to leverage layer cache
COPY backend/go.mod backend/go.sum ./
# web/ is prebuilt and embedded into the binary via //go:embed;
# to rebuild the frontend inside the image, add a node stage here:
#   FROM node:20-bookworm AS web
#   COPY frontend-web/ /web
#   RUN cd /web && npm ci && npm run build
# then copy the output into backend/web.
COPY backend/ ./

ENV CGO_ENABLED=1 GOPROXY=https://goproxy.cn,direct
RUN go build -o /out/picooffice . \
 && go build -o /out/picobackup ./cmd/picobackup

# ---------- runtime ----------
FROM debian:bookworm-slim

# libreoffice provides soffice; fonts-noto-cjk keeps CJK exports readable
RUN apt-get update && apt-get install -y --no-install-recommends \
    libreoffice \
    fonts-noto-cjk \
    fonts-noto-cjk-extra \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/picooffice /app/picooffice
COPY --from=builder /out/picobackup /app/picobackup

ENV PICO_PORT=8080 \
    PICO_DB=/data/db/picooffice.db \
    PICO_STORAGE=/data/storage \
    PICO_EXPORT_WORKERS=2

EXPOSE 8080

# data dirs for volumes
RUN mkdir -p /data/db /data/storage

ENTRYPOINT ["/app/picooffice"]
