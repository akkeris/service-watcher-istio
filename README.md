# service-watcher-istio

Watches for new Kubernetes services for Akkeris apps, and creates Istio VirtualService entries to expose them outside the cluster.

## Environment Variables Descriptions

- `NAMESPACE_BLACKLIST`: Do not create virtualservices for services in any namespace in this list. Should be a comma-separated string.
- `IGNORE_LABELS`: Do not create virtualservices for any service with a label on this list. Should be a comma-separated string.

## Example/Suggested Environment Variables

This should give you a good starting point. If you find that undesired virtualservices are still being made, you can expand these lists with other namespaces or labels.

**NAMESPACE_BLACKLIST**

`flux,cattle-system,cattle-prometheus,kube-system,kube-public,testcafe,akkeris-system,istio-system,nginx-ingress-i,prometheus,sites-system,velero`

**IGNORE_LABELS**

`akkeris.io/container-ports`