# Build a No-Op Function

We will start with a no-op function that does nothing but prints the data it
receives from Crossplane. This function will be used as a base for all other
functions we will write in this tutorial.

## Building the Function

Initialize the go module.
```bash
mkdir xfn-noop
cd xfn-noop
```
```bash
# Do not forget to change "muvaf" github namespace to your own.
go mod init github.com/crossplane-contrib/contribfest/lab-composition-functions/xfn-noop
```

Create a `main.go` file that only prints the standard input it receives.
```go
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(b))
}
```

Create a `Dockerfile` that builds an image for our Go program to be used as
a function.
```dockerfile
FROM golang:1.20-alpine3.17 as builder

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /function .

###
FROM alpine:3.17.3

COPY crossplane.yaml /crossplane.yaml
COPY --from=builder /function /function

ENTRYPOINT ["/function"]
```

Lastly, we will add a `crossplane.yaml` file with metadata about the function.
```yaml
apiVersion: meta.pkg.crossplane.io/v1alpha1
kind: Function
metadata:
  name: xfn-noop
  annotations:
    meta.crossplane.io/maintainer: ContribFest Crossplane
    meta.crossplane.io/source: github.com/crossplane-contrib/contribfest/lab-composition-functions/
    meta.crossplane.io/license: Apache-2.0
    meta.crossplane.io/description: |
      A Composition Function that prints the data it receives and returns a no-op
      result to Crossplane.
```

Let's build and push the image to a registry. Assumes you are already logged in
to DockerHub.
```bash
docker build --tag muvaf/xfn-noop:v0.1.0 .
docker push muvaf/xfn-noop:v0.1.0
```

Let's test it locally. The following is an example `FunctionIO` that we can get
from Crossplane. Create a file called `test.yaml` with the following content:
```yaml
apiVersion: apiextensions.crossplane.io/v1alpha1
kind: FunctionIO
observed:
  composite:
    resource:
      apiVersion: contribfest.crossplane.io/v1alpha1
      kind: XRobotGroup
      metadata:
        name: somename
desired:
  composite:
    resource:
      apiVersion: contribfest.crossplane.io/v1alpha1
      kind: XRobotGroup
      metadata:
        name: somename
  resources:
    - name: one-robot
      resource:
        apiVersion: dummy.upbound.io/v1beta1
        kind: Robot
        spec:
          forProvider:
            color: yellow
```

Now let's give it to our composition function.
```bash
cat test.yaml | docker run -i --rm muvaf/xfn-noop:v0.1.0
```

Alternatively, we can run the Go program directly.
```bash
cat test.yaml | go run main.go
```

## Installation

Now that we have the function image, we can use it in a Composition. Assuming
you've gone through the [pre-requisites](01-prerequisites.md), you should have a
kind cluster with Crossplane installed with composition functions feature flag
enabled.

Let's create a simple `CompositeResourceDefinition` to have an API called
`RobotGroup`.
```bash
cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: xrobotgroups.contribfest.crossplane.io
spec:
  group: contribfest.crossplane.io
  claimNames:
    kind: RobotGroup
    plural: robotgroups
  names:
    kind: XRobotGroup
    plural: xrobotgroups
  versions:
    - name: v1alpha1
      served: true
      referenceable: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                count:
                  type: number
EOF
```

Create a `Composition` that uses the `xfn-noop` function.
```bash
cat <<EOF | kubectl apply -f -
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: robotgroup
spec:
  compositeTypeRef:
    apiVersion: contribfest.crossplane.io/v1alpha1
    kind: XRobotGroup
  resources:
    - name: one-robot
      base:
        apiVersion: iam.dummy.upbound.io/v1alpha1
        kind: Robot
        spec:
          forProvider:
            color: yellow
  functions:
  - name: my-noop-function
    type: Container
    container:
      image: muvaf/xfn-noop:v0.1.0
EOF
```

## Usage

Let's create an instance of `RobotGroup` and see what happens.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: contribfest.crossplane.io/v1alpha1
kind: RobotGroup
metadata:
  name: my-robot-group
spec: {}
EOF
```

We should see that a single `Robot` resource is created as defined under `resources`
array of our `Composition`.
```bash
kubectl get robot -o yaml
```

As expected, our function printed the input as is without manipulating anything
so we get exactly what's defined in the `resources` array.

### Cleanup

Delete the `RobotGroup` instance.
```bash
kubectl delete robotgroup my-robot-group
```