1. To install the Timber Docker logging plugin on a host:

```bash
docker plugin install timberio/docker-logging-driver:0.1.0 --alias timber
```

2. To use the Timber Docker logging driver on a single container, run the container with the following options:

```bash
docker run --log-driver timber \
  --log-opt timber-api-key="{{my-timber-api-key}}" \
  CONTAINER_NAME
```

### OR

To use the Timber Docker logging driver for all containers on a running on a host, you will need to [configure the
Docker daemon](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file).

Assuming the host system is Linux:

2. Add the necessary configuration to `/etc/docker/daemon.json` replacing `my-timber-api-key` with your API key

    ```json
    {
      "log-driver": "timber",
      "log-opts": {
        "timber-api-key": "{{my-timber-api-key}}"
      }
    }
    ```

3. (Re)start the Docker daemon

    ```bash
    sudo service docker restart
    ```

4. (Re)start Docker containers
