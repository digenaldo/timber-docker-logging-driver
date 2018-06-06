# Configuration

The Timber Docker logging driver exposes its configuration options Docker options.

These options can be set with the Docker `--log-opt` cli flag or the `log-opts` json key in the [Docker daemon
configuration file](https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-configuration-file).

## Configuration Options

| Option          | Required? | Default Value | Description    |
| --------------- | :-------: | :-----------: | -------------- |
| `timber-api-key`| yes       |               | Timber API key |
