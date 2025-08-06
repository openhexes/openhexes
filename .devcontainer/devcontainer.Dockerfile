FROM node:24-bookworm

ENV DEVCONTAINER="true" \
    CGO_ENABLED="0" \
    PATH="/root/.local/bin:/usr/local/go/bin:/root/go/bin:${PATH}"

RUN apt -y update && apt install -y --no-install-recommends \
    gettext \
    net-tools \
    postgresql-client \
    psmisc \
    telnet \
    && rm -rf /var/lib/apt/lists/*

RUN curl -sSfL https://raw.githubusercontent.com/ajeetdsouza/zoxide/main/install.sh | sh

