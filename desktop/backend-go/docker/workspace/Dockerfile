# Lightweight workspace container for BusinessOS
FROM alpine:3.19

# Install essential tools in a single layer to keep image small
RUN apk add --no-cache \
    bash \
    zsh \
    git \
    curl \
    vim \
    nano \
    python3 \
    py3-pip \
    nodejs \
    npm \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user with UID 1000
RUN addgroup -g 1000 workspace && \
    adduser -D -u 1000 -G workspace -s /bin/bash workspace

# Set working directory
WORKDIR /workspace

# Change ownership of workspace directory
RUN chown -R workspace:workspace /workspace

# Switch to non-root user
USER workspace

# Set default shell
ENV SHELL=/bin/bash

# Add common aliases and environment setup
RUN echo 'alias ll="ls -lah"' >> /home/workspace/.bashrc && \
    echo 'alias la="ls -A"' >> /home/workspace/.bashrc && \
    echo 'alias l="ls -CF"' >> /home/workspace/.bashrc && \
    echo 'export PS1="\[\033[01;32m\]\u@\h\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ "' >> /home/workspace/.bashrc

# Default command
CMD ["/bin/bash"]
