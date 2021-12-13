# cloud-ready-checks
Readiness (await) logic for cloud resources

This is a library of cloud await/readiness checks, based on some code from the [native Kubernetes provider](https://github.com/pulumi/pulumi-kubernetes). 
The idea is to make it easier to write and test await logic for cloud resources (not just Kubernetes).

## Repo layout

- internal - test data to validate the state checkers
- pkg/checker - a generic state checker
- pkg/kubernetes - Kubernetes-specific state checks
