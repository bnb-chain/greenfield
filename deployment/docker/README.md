![Markdown](https://img.shields.io/badge/Make-4.3-red)
[![Markdown](https://img.shields.io/badge/Golang-1.18-green)](https://go.dev/dl/)
[![Markdown](https://img.shields.io/badge/Docker-23.0.5-blue)](https://docs.docker.com/get-docker/)
[![Markdown](https://img.shields.io/badge/Docker%20Compose-1.29.2-blue)](https://github.com/docker/compose)
![Markdown](https://img.shields.io/badge/jq-1.6-red)

## Minimum System Requirements

- 1 TB of free disk space, accessible at a minimum read/write speed of 100 MB/s.
- 4 cores of CPU and 8 GB of memory (RAM).
- A broadband Internet connection with upload/download speeds of at least 1 MB/s.

## Setting up a working envrionment

| Please check the [greenfield repo](https://github.com/bnb-chain/greenfield) for information on the testnet, including the correct version of the binaries to use and details about the config file. |
| --- |

### 1. Make a working directory $NODE_HOME (i.e. ~/.gnfd)

```bash
$ mkdir ~/.gnfd
$ mkdir ~/.gnfd/config
$ mkdir ~/.gnfd/data
```

### 2. Download testnet configuration files from https://github.com/bnb-chain/greenfield/releases and copy them into $NODE_HOME/config

```bash
$ wget  $(curl -s https://api.github.com/repos/bnb-chain/greenfield/releases/latest |grep browser_ |grep testnet_config |cut -d\" -f4)
$ unzip testnet_config.zip
$ cp testnet_config/*  ~/.gnfd/config/
```

You can edit the node's moniker at $NODE_HOME/config/config.toml file:

```bash
# A custom human readable name for this node
moniker = "<your_custom_moniker>"
```

### 3. docker-compose up

```bash
$ docker-compose up
```

---
**Tips** 

```bash
# Assigning permissions to config and data folders.
$ chmod 777 -R ~/.gnfd/config
$ chmod 777 -R ~/.gnfd/data
```

---