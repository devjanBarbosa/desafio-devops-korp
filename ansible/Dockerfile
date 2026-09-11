FROM alpine:3.20

RUN apk add --no-cache \
    python3 \
    py3-pip \
    curl \
    docker-cli \
    docker-cli-compose \
    gcc \
    python3-dev \
    musl-dev \
    libffi-dev

RUN python3 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"
ENV DOCKER_API_VERSION="1.45"

RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir \
    ansible-core==2.16.4 \
    docker==7.1.0 \
    requests

RUN ansible-galaxy collection install community.docker:3.10.0

WORKDIR /workspace/ansible

ENTRYPOINT ["ansible-playbook"]