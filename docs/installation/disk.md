1. Find the Docker log path

    _On linux systems, the default log location for a container is `/var/lib/docker/containers/<container id>/<container id>-json.log`_

    ```bash
    docker inspect --format='{{.LogPath}}' $INSTANCE_ID
    ```

    ```text
    /var/lib/docker/containers/\
    e48d96758131806372b694199e6eac78f4d56d3d504e92e60e763b75d8de593f/\
    e48d96758131806372b694199e6eac78f4d56d3d504e92e60e763b75d8de593f-json.log
    ```

    `$INSTANCE_ID` is the id of a container not using a custom log path.

2. Configure the Timber Agent to watch all files on the Docker log path

    _Options for the Timber Agent config file can be found [here](https://docs.timber.io/platforms/other/timber-agent/configuration-file/)._

    **Timber Agent config file (timber.toml)**

    ```toml
    default_api_key = "TIMBER_API_KEY"

    [[files]]
    path = "/var/log/containers/**/*.log"
    ```
3. Run Timber Agent container

    ```bash
    docker run \
      -v $DOCKER_LOG_PATH:/var/log/containers \
      -v $TIMBER_TOML_PATH:/etc/timber.toml \
      timberio/agent:0.9.2 capture-files
    ```

    `$DOCKER_LOG_PATH` is the path where Docker container log files are stored.

    `$TIMBER_TOML_PATH` is the path on disk to Timber Agent configuration file.
