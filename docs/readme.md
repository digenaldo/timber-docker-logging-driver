---
title: Docker Logs
items:
 - installation
 - configuration.md
 - troubleshooting.md
---

# Docker Logging Driver

With the [Timber Docker logging driver](https://github.com/timberio/timber-docker-logging-driver) you can leverage the [Docker daemon](https://docs.docker.com/config/containers/logging/configure/) to its fullest. Let the daemon handle shipping while Timber handles utilization.

# Docker Logs from Disk

Assuming you are using the [Docker default JSON logging driver](https://docs.docker.com/config/containers/logging/json-file/), you can ship your
containers logs by running the [Timber Agent](https://github.com/timberio/agent) container.