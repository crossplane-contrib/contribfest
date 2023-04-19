# Prerequisites

* `kubectl`
* `docker`
* `go`

## Preparing the environment

1. Create a kind cluster. See [kind documentation](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
   for installation instructions.
   ```bash
   kind create cluster --wait 5m
   kubectl create namespace crossplane-system
   ```
2. Log in to DockerHub and add your credentials to the cluster so that Crossplane
   does not hit pull limits when it pulls your function images.
   ```bash
   docker login
   ```
   ```bash
   kubectl -n crossplane-system create secret generic dockerhub \
     --from-file=.dockerconfigjson=$HOME/.docker/config.json \
     --type=kubernetes.io/dockerconfigjson
   ```
   Additionally, export your username as an environment variable so that we can
    use it later.
    ```bash
    export REGISTRY=<your-dockerhub-username>
    ```
3. Install Crossplane v1.11.0 or later installed with composition functions
   feature flag enabled.
   ```bash
   helm install crossplane --namespace crossplane-system crossplane-stable/crossplane \
     --create-namespace --wait \
     --set "args={--debug,--enable-composition-functions}" \
     --set "xfn.enabled=true" \
     --set "xfn.args={--debug}" \
     --set "imagePullSecrets={dockerhub}"
   ```
4. Install the `provider-dummy` package. It will manage external resources in a
   local server that does not require authentication.
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: pkg.crossplane.io/v1
   kind: Provider
   metadata:
     name: provider-dummy
   spec:
     package: xpkg.upbound.io/upbound/provider-dummy:v0.2.0
   EOF
   ```
   Deploy its dummy server as well.
   ```bash
   kubectl -n crossplane-system apply -f https://raw.githubusercontent.com/upbound/provider-dummy/v0.2.0/cluster/server-deployment.yaml
   ```
   Configure it to talk to the server we just deployed.
   ```bash
   cat <<EOF | kubectl apply -f -
   apiVersion: dummy.upbound.io/v1alpha1
   kind: ProviderConfig
   metadata:
     name: default
   spec:
     endpoint: http://server-dummy.crossplane-system.svc.cluster.local
   EOF
   ```

Done!