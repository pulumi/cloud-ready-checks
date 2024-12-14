CHANGELOG
=========

## Unreleased

## 1.2.0 (2024-12-11)

### Added

- Pod warnings and errors now include a container's termination state and
  message, if present. By default the termination message is read from
  `/dev/termination-log` but can be configured with `terminationMessagePath` or
  `terminationMessagePolicy`. (https://github.com/pulumi/cloud-ready-checks/pull/17)

### Changed

- Upgrade Kubernetes client libraries to v0.32.0 (https://github.com/pulumi/cloud-ready-checks/pull/61)
- Upgrade Go to v1.23 (https://github.com/pulumi/cloud-ready-checks/pull/54)

## 1.1.0 (2023-12-13)

- Upgrade Go to v1.21 (https://github.com/pulumi/cloud-ready-checks/pull/5)
- Upgrade pulumi/pulumi to v3.96.2 (https://github.com/pulumi/cloud-ready-checks/pull/6)
- Upgrade Kubernetes client libraries to v0.29.0 (https://github.com/pulumi/cloud-ready-checks/pull/4)

## 1.0.0 (2022-01-04)

Initial release
