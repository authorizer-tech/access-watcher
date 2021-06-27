# access-watcher

[![Latest Release](https://img.shields.io/github/v/release/authorizer-tech/access-watcher)](https://github.com/authorizer-tech/access-watcher/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/authorizer-tech/access-watcher)](https://goreportcard.com/report/github.com/authorizer-tech/access-watcher)
[![Slack](https://img.shields.io/badge/slack-%23authorizer--tech-green)](https://authorizer-tech.slack.com)

An access-watcher serves Watch RPCs that stream changes from a relation tuple [changelog](https://authorizer-tech.github.io/docs/overview/architecture#changelog) in near real-time to clients interested in changes to one or more [namespaces](https://authorizer-tech.github.io/docs/overview/concepts/namespaces).

An instance of an `access-watcher` is similar to the `watchserver` implementation called out in the [Google Zanzibar](https://research.google/pubs/pub48190/) paper (see Section 3).