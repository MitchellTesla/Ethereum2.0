# Consul Terraform Sync

Consul Terraform Sync (just Sync from here on) is a service-oriented tool for managing network infrastructure near real-time. Sync runs as a daemon and integrates the network topology maintained by your Consul cluster with your network infrastructure to dynamically secure and connect services.

* Website: [consul.io/docs/nia](https://www.consul.io/docs/nia)

## Community Support
If you have questions about how `consul-terraform-sync` works, its capabilities or anything other than a bug or feature request (use github's issue tracker for those), please see our community support resources.

Community portal: [discuss.hashicorp.com](https://discuss.hashicorp.com/tags/c/consul/29/consul-terraform-sync)

Other resources: [consul.io/community](https://www.consul.io/community)

Additionally, for issues and pull requests, we'll be using the 👍 reactions as a rough voting system to help gauge community priorities. So please add 👍 to any issue or pull request you'd like to see worked on. Thanks.

## Installation
Sync is a daemon that runs alongside [Consul](https://github.com/hashicorp/consul), similar to other Consul ecosystem tools like [Consul Template](https://github.com/hashicorp/consul-template). Sync is not included with the Consul binary and will need to be installed separately.

### Download
To install Sync, find the appropriate package for your system and download it as a zip archive. Unzip the package to extract the binary named `consul-terraform-sync`. Move the consul-terraform-sync binary to a location available on your `$PATH`.

  1. Download a pre-compiled, released version from the [Sync release page](https://releases.hashicorp.com/consul-terraform-sync/).
  1. Extract the binary using `unzip` or `tar`.
  1. Move the binary into `$PATH`.

```shell
$ wget https://releases.hashicorp.com/consul-terraform-sync/${VERSION}/consul-terraform-sync_${VERSION}_${OS}_${ARCH}.zip
$ unzip consul-terraform-sync_${VERSION}_${OS}_${ARCH}.zip
$ mv consul-terraform-sync /usr/local/bin/consul-terraform-sync
```

### Build from Source

You can also build Sync from source.

  1. Clone the repository to your local machine.
  1. Pick a [version](https://github.com/hashicorp/consul-terraform-sync/releases) or build from master.
  1. Build Sync using the [Makefile](Makefile).
  1. The `consul-terraform-sync` binary is now installed to `$GOPATH/bin`.

```shell
$ git clone https://github.com/hashicorp/consul-terraform-sync.git
$ cd consul-terraform-sync
$ git fetch --all
$ git checkout tags/vX.Y.Z
$ make dev
$ which consul-terraform-sync
```

### Verify

Once installed, verify the installation works by prompting the help option.

```shell
$ consul-terraform-sync -h
Usage of consul-terraform-sync:
  -config-dir value
      A directory to load files for configuring Sync. Configuration files
      require an .hcl or .json file extention in order to specify their format.
      This option can be specified multiple times to load different directories.
  -config-file value
      A file to load for configuring Sync. Configuration file requires an
      .hcl or .json extension in order to specify their format. This option can
      be specified multiple times to load different configuration files.
  -once
      Render templates and run tasks once. Does not run the process as a daemon
      and disables wait timers.
  -version
      Print the version of this daemon.
```

## Configuration

[Documentation to configure Sync](https://consul.io/docs/nia/configuration)

# Build-status (remote) ' in progress '

Network infrastructure and (remote) custom build and integration in development...
