# Kuberhealthy DNS Resolution Check

This check verifies that DNS resolution works inside and outside the cluster by resolving a configured hostname. It can optionally resolve through DNS service endpoints when a label selector is provided.

## What It Does

1. Resolves the configured hostname directly.
2. Optionally queries DNS service endpoints when `DNS_POD_SELECTOR` is set.
3. Reports failure when lookups fail or time out.

## Configuration

All configuration is controlled via environment variables.

- `HOSTNAME`: Hostname to resolve (required).
- `NODE_NAME`: Node name for readiness checks (required).
- `NAMESPACE`: Namespace for DNS endpoints (optional).
- `DNS_POD_SELECTOR`: Label selector for DNS endpoints (optional).

Kuberhealthy injects these variables automatically into the check pod:

- `KH_REPORTING_URL`
- `KH_RUN_UUID`
- `KH_CHECK_RUN_DEADLINE`

## Build

Use the `Justfile` to build or test the check:

```bash
just build
just test
```

## Example HealthCheck (Service DNS)

```yaml
apiVersion: kuberhealthy.github.io/v2
kind: HealthCheck
metadata:
  name: dns-status-internal
  namespace: kuberhealthy
spec:
  runInterval: 2m
  timeout: 15m
  podSpec:
    spec:
      containers:
        - name: dns-resolution-check
          image: kuberhealthy/dns-resolution-check:sha-<short-sha>
          imagePullPolicy: IfNotPresent
          env:
            - name: HOSTNAME
              value: "kubernetes.default"
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          resources:
            requests:
              cpu: 10m
              memory: 50Mi
      restartPolicy: Never
```

## Example HealthCheck (Endpoint DNS)

```yaml
apiVersion: kuberhealthy.github.io/v2
kind: HealthCheck
metadata:
  name: dns-status-endpoint
  namespace: kuberhealthy
spec:
  runInterval: 2m
  timeout: 15m
  podSpec:
    spec:
      containers:
        - name: dns-resolution-check
          image: kuberhealthy/dns-resolution-check:sha-<short-sha>
          imagePullPolicy: IfNotPresent
          env:
            - name: HOSTNAME
              value: "kubernetes.default"
            - name: NAMESPACE
              value: "kube-system"
            - name: DNS_POD_SELECTOR
              value: "k8s-app=kube-dns"
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
      restartPolicy: Never
```

A full install bundle with RBAC is available in `healthcheck.yaml`.

## Image Tags

- `sha-<short-sha>` tags are published on every push to `main`.
- `vX.Y.Z` tags are published when a matching Git tag is pushed.
