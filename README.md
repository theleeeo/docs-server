# Docs-Server

This is a server for polling git repositories for [swagger](https://swagger.io/specification/) and serving them using [redoc](https://redocly.com/).

## Table of Contents

- [Installation](#installation)
- [Config](#configuration)

## Installation

The easiest way to install is to use the docker image. Make sure to replace `tag` with the version you want to use.

```bash
$ docker run -p 4444:4444 -v ./config.yml:/app/config.yml ghcr.io/theleeeo/docs-server:tag
```

You can also build from source.

```bash
$ git clone https://github.com/theleeeo/docs-server
$ cd yourproject
# Add your config.yml
$ go run .
```

## Configuration

The configuration is done using a `config.yml` file in the root of the project.

The following is an example of a config file.

```yml
# The minimum log level to log
# Possible values are: debug, info, warn, error
# Default is info
log_level: info

provider:
  github:
    # The user or organization that the repo belongs to
    owner: theleo
    # The name of the repo
    repo: a-swagger-repo
    # The path to look for swagger files in relative to the root of the repo
    path_prefix: api/
    # The suffix of the swagger files
    file_suffix: .swagger.json
    # The maximum number of tags to show as versions
    max_tags: 10
    # The github token to use for the client
    # This is to allow a higher rate limit
    # Using this to access private repos will not work since the app will not be able to access it anyways
    auth_token: github_pat_SuperSecretToken

server:
  # How often should the server poll the provider for new vesions
  poll_interval: 30m
  # Should the server act as a proxy, fecthing the swagger files from the provider and serving them
  # This is useful if the provider is not accessible from the internet or requires authentication
  proxy: false

app:
  # The host:port to run the server on
  address: localhosts:4444
  # An optional prefix to the path that the app listens on
  # This is useful if you are running the app behind a reverse proxy
  path_prefix: /docs

design:
  # The title that will be shown in the header
  header_name: Leo Evil Inc'
  # A filepath or url to the logo that will be shown in the header
  # Files should be placed in the ./public folder
  header_image: https://theleo.se/favicon.png
  # The color of the header
  # Allowed values are any valid css color (hex, rgb, named colors, etc.)
  # Default is the color of the website background
  header_color: "#000000"
  # The filepath or url to the icon that will be shown in the browser tab
  # Files should be placed in the ./public folder
  favicon: https://theleo.se/favicon.png
```