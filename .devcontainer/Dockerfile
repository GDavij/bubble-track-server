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

RUN if id "vibecoder" >/dev/null 2>&1; then \
      echo "vibecoder user already exists, skipping"; \
    else \
      useradd -m -s /bin/bash vibecoder && \
      usermod -aG sudo vibecoder && \
      echo "vibecoder ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers; \
    fi

RUN go install github.com/air-verse/air@latest && \
    mv /go/bin/air /usr/local/bin/air

USER vibecoder
WORKDIR /home/vibecoder/app

ENV SHELL=/bin/bash

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]