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
provider:
  github:
    # The user or organization that the repo belongs to
    owner: theleo
    # The name of the repo
    repo: a-swagger-repo
    # The maximum number of tags to show as versions
    max_tags: 10
    # The github token to use for the client
    # This is to allow a higher rate limit
    # Using this to access private repos will not work since the app will not be able to access it anyways
    auth_token: github_pat_SuperSecretToken

server:
  # How often should the server poll the provider for new vesions
  poll_interval: 30m
  # The path to look for swagger files in relative to the root of the repo
  path_prefix: api/
  # The suffix of the swagger files
  file_suffix: .swagger.json

app:
  # The host:port to run the server on
  address: localhosts:4444

design:
  # The title that will be shown in the header
  company_name: Leo Evil Inc'
  # A filepath or url to the logo that will be shown in the header
  company_logo: https://theleo.se/favicon.png
```