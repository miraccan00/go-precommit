# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache dependency downloads separately from the build.
COPY go.mod go.sum ./
RUN go mod download

# Build a fully static binary (no libc dependency).
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" \
      -o /out/go-precommit \
      .

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
# alpine:3 keeps the image small (~12 MB) while providing a proper filesystem
# (needed for volume mounts, timezone data, and CA certificates).
FROM alpine:3

# git is required by hooks that shell out to the git CLI
# (check-added-large-files, no-commit-to-branch, destroyed-symlinks, etc.)
RUN apk add --no-cache ca-certificates git tzdata \
    && mkdir -p /workspace

COPY --from=builder /out/go-precommit /usr/local/bin/go-precommit

# Mount your project directory here:
#
#   docker run --rm \
#     -v "$(pwd)":/workspace \
#     -w /workspace \
#     ghcr.io/miraccan00/go-precommit run --all-files
#
WORKDIR /workspace

ENTRYPOINT ["go-precommit"]
CMD ["run"]
