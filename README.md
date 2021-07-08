# access-watcher

[![Latest Release](https://img.shields.io/github/v/release/authorizer-tech/access-watcher)](https://github.com/authorizer-tech/access-watcher/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/authorizer-tech/access-watcher)](https://goreportcard.com/report/github.com/authorizer-tech/access-watcher)
[![Slack](https://img.shields.io/badge/slack-%23authorizer--tech-green)](https://authorizer-tech.slack.com)

An access-watcher serves Watch RPCs that stream changes from a relation tuple [changelog](https://authorizer-tech.github.io/docs/overview/architecture#changelog) in near real-time to clients interested in changes to one or more [namespaces](https://authorizer-tech.github.io/docs/overview/concepts/namespaces).

An instance of an `access-watcher` is similar to the `watchserver` implementation called out in the [Google Zanzibar](https://research.google/pubs/pub48190/) paper (see Section 3).

# Getting Started
If you want to setup an instance of the Authorizer platform as a whole, browse the API References, or just brush up on the concepts and design of the platform, take a look at the [official platform documentation](https://authorizer-tech.github.io/docs/overview/introduction). If you're only interested in running the access-watcher then continue on.

## Setup and Installation
> An access-watcher is not a standalone application. It is intended to be deployed alongside an existing [access-controller](https://github.com/authorizer-tech/access-controller) deployment. Setting up an access-watcher will fail if an existing access-controller has not been deployed.

### Pre-compiled Binaries
Download the [latest release](https://github.com/authorizer-tech/access-watcher/releases) an extract it.

```console
$ ./bin/access-watcher -config <config-path> -grpc-port 50052
```

## Next Steps...
Take a look at the official [Documentation](https://authorizer-tech.github.io/docs/overview/introduction), [API Reference](https://authorizer-tech.github.io/docs/api-reference/overview) and [Examples](https://authorizer-tech.github.io/docs/overview/examples/examples-intro).

# Community
The access-watcher is an open-source project and we value and welcome new contributors and members
of the community. Here are ways to get in touch with the community:

* Slack: [#authorizer-tech](https://authorizer-tech.slack.com)
* Issue Tracker: [GitHub Issues](https://github.com/authorizer-tech/access-watcher/issues)