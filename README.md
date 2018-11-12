# Timber Docker logging driver

[![Built by Timber.io](https://res.cloudinary.com/timber/image/upload/v1503615886/built_by_timber_wide.png)](https://timber.io/?utm_source=github&utm_campaign=timberio%2Fagent)

[![GitHub release](https://img.shields.io/github/release/timberio/timber-docker-logging-driver.svg)](https://github.com/timberio/timber-docker-logging-driver/releases/latest)
[![license](https://img.shields.io/github/license/timberio/timber-docker-logging-driver.svg)](https://github.com/timberio/timber-docker-logging-driver/blob/master/LICENSE)

The Timber Docker logging driver is a Docker logging plugin for collecting Docker container logs and shipping them to
Timber.io.

_Only container logs on stdout and stderr will be collected and shipped._

1. [**Installation**](#installation)
1. [**Usage**](#usage)
1. [**Configuration**](#configuration)
1. [**Contributing**](#contributing)

## Installation

To install the Timber Docker logging plugin on a host:

```bash
docker plugin install timberio/docker-logging-driver:0.1.1 --alias timber
```

To verify the plugin is installed and enabled:

```bash
docker plugin ls
```

```text
ID                  NAME                DESCRIPTION             ENABLED
64dc540271a0        timber:latest       Timber logging driver   true
```

### Uninstall

To remove the Timber Docker loggin plugin on a host:

```bash
docker plugin disable timber
docker plugin rm timber
```

## Usage

Docker logging drivers can be used in two ways, on a per container basis or for every container on a host.

_Both options require a Timber API key.
If you need a key go [here](https://timber.io/docs/app/applications/obtaining-api-key)._

### Per Container

To use the Timber Docker logging driver on a single container, run the container with the following options:

```bash
docker run --log-driver timber \
  --log-opt timber-api-key="{{my-timber-api-key}}" \
  CONTAINER_NAME
```

Here `my-timber-api-key` is your Timber API key and `CONTAINER_NAME` is the container to run.

### Per Host

To use the Timber Docker logging driver for all containers on a running on a host, you will need to [configure the
Docker daemon](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file).

Assuming the host system is Linux:

1. Add the necessary configuration to `/etc/docker/daemon.json` replacing `my-timber-api-key` with your API key

    ```json
    {
      "log-driver": "timber",
      "log-opts": {
        "timber-api-key": "{{my-timber-api-key}}"
      }
    }
    ```

1. Restart the Docker daemon

    ```bash
    sudo service docker restart
    ```

## Configuration

These are the supported Timber Docker logging driver options, and can be set with the Docker `--log-opt` flag.

| Option          | Required? | Default Value | Description    |
| --------------- | :-------: | :-----------: | -------------- |
| `timber-api-key`| yes       |               | Timber API key |

## Contributing

This project uses Dep as the dependency manager and all vendorized dependencies are committed into version control.
If you make a change that includes a new dependency, please make sure to add it to the dependency manager properly. You
can do this by editing the Gopkg.toml file in the root of the project (format documentation). After editing the file,
run dep ensure to update the vendor folder.

It is recommended to test the plugin on a docker-machine environment or one where you can access the docker logs easily.
Verbose logging can also be turned on by setting the `LOG_LEVEL` environment variable to `debug`.

To build and install the plugin locally from source:

```bash
./plugin install
```

```text
ID                  NAME                DESCRIPTION             ENABLED
1d046b706212        timber:latest       Timber logging driver   true
```

To run a container with the plugin:

```bash
docker run --log-driver timber \
  --log-opt timber-api-key="{{my-timber-api-key}}" \
  --env LOG_LEVEL="debug"
  CONTAINER_NAME
```

To uninstall the plugin locally:

```bash
./plugin uninstall
```
