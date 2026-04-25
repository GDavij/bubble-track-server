FROM golang:1.26.2-bookworm

ENV DEBIAN_FRONTEND=noninteractive
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility

RUN apt-get update && apt-get install -y \
    sudo \
    curl \
    git \
    wget \
    htop \
    pciutils \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN go install github.com/air-verse/air@latest && \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest

RUN if id "vibecoder" >/dev/null 2>&1; then \
      echo "vibecoder user already exists, skipping"; \
    else \
      useradd -m -s /bin/bash -u 1000 -G sudo vibecoder && \
      echo "vibecoder ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers; \
    fi

RUN chown -R vibecoder:vibecoder /go

USER vibecoder

ENV SHELL=/bin/bash