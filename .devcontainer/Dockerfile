FROM mcr.microsoft.com/devcontainers/go:1-1.25-bookworm

# Linux clipboard binaries — exercised by internal/clipboard tests on Linux CI.
# xclip for X11, wl-clipboard for Wayland sessions.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        xclip \
        wl-clipboard \
    && rm -rf /var/lib/apt/lists/*

# goreleaser via the official apt repository (pinned to v2.x line via the repo).
RUN echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' \
        > /etc/apt/sources.list.d/goreleaser.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends goreleaser \
    && rm -rf /var/lib/apt/lists/*

# golangci-lint via the official install script. Kept unpinned: the script
# resolves the latest tagged release, which matches what the user would get
# in CI by default. Pin a specific version here if reproducibility matters.
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
        | sh -s -- -b /usr/local/bin
