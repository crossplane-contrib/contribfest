# Building a Function that Sets Random Value

With composition, there is no way to set a random value to a field because every
time it reconciles, it'll override the value that was set in the previous
reconciliation pass. In this tutorial, we'll create a function that parses all
`Robot` objects in the desired state and sets a random color to them if they
don't have one because it is a required parameter.

We will use DockerHub to push & pull the images we build. The following environment
variable will be required to make sure the commands work with your own images.
```bash
export REGISTRY=<your-dockerhub-username>
```

Let's build on top of our no-op function.
```bash
cp -a xfn-noop xfn-random
cd xfn-random
```

Change the function name to `xfn-random` in all files.
```bash
# On Mac
sed -i '' 's/xfn-noop/xfn-random/g' *
# On Linux
sed -i 's/xfn-noop/xfn-random/g' *
```

NOTE: The rest of the guide assumes that you already have the `CompositeResourceDefinition`
and `Composition` created from the previous tutorial. If you don't, you can
go back [installation section](02-xfn-noop.md#installation) and create them.

### Parsing Input

Our function currently does nothing, hence it doesn't need to parse the input it
receives. We will first need to parse the input as proper objects so that we can
set values.

Let's import the type of the input object.
```bash
go get github.com/crossplane/crossplane
go get sigs.k8s.io/yaml
```

Let's update the `main.go` file to parse the input as a `FunctionIO` object but
still do nothing.
```go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	"sigs.k8s.io/yaml"
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v", err)
		os.Exit(1)
	}
	obj := &v1alpha1.FunctionIO{}
	if err := yaml.Unmarshal(b, obj); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal stdin: %v", err)
		os.Exit(1)
	}
	result, err := yaml.Marshal(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(result))
}
```

```bash
# This is to make sure go.sum is tidied up after all the go get commands.
go mod tidy
```

Run this program with our existing `test.yaml` to see it's working.
```bash
cat test.yaml | go run main.go
```

Let's define an array of colors that we'll use to set the random value.
```go
var (
Colors = []string{"red", "green", "blue", "yellow", "orange", "purple", "black", "white"}
)
```

### Manipulating the Desired State

Now, we have a `FunctionIO` object that contains the desired and observed `Robot`
objects. We will make sure every element in the `desired.resources` array has its
color set to a value; observed or random.

Our function will work on only the `Robot` objects so let's import the necessary
machinery.
```bash
go get github.com/upbound/provider-dummy
go get k8s.io/apimachinery/pkg/runtime
go get k8s.io/apimachinery/pkg/util/json
go get sigs.k8s.io/yaml
```

Right after we get the `FunctionIO` object, here we extract the colors of the all
observed `Robot`s to make sure to propagate them to the desired state so that
they don't get overridden.
```go
    colors := map[string]string{}
    for _, observed := range obj.Observed.Resources {
        r := &dummyv1alpha1.Robot{}
        if err := json.Unmarshal(observed.Resource.Raw, r); err != nil {
            fmt.Fprintf(os.Stderr, "failed to unmarshal observed resource: %v", err)
            os.Exit(1)
        }
        colors[observed.Name] = r.Spec.ForProvider.Color
    }
```

In the next loop, we skip all the entries that already have a color set and
generate a random color for the rest.

```go
    for i, desired := range obj.Desired.Resources {
        r := &dummyv1alpha1.Robot{}
        if err := yaml.Unmarshal(desired.Resource.Raw, r); err != nil {
            fmt.Fprintf(os.Stderr, "failed to unmarshal desired resource: %v", err)
            os.Exit(1)
        }
        if colors[desired.Name] != "" {
            r.Spec.ForProvider.Color = colors[desired.Name]
        } else {
            r.Spec.ForProvider.Color = Colors[rand.Intn(len(Colors))]
        }
        // NOTE: We need to use a JSON marshaller here because runtiem.RawExtension
        // type expects a JSON blob.
        raw, err := json.Marshal(r)
        if err != nil {
            fmt.Fprintf(os.Stderr, "failed to marshal resource: %v", err)
            os.Exit(1)
        }
        obj.Desired.Resources[i].Resource = runtime.RawExtension{Raw: raw}
    }
```

Here is the full `main.go` file.
<details>
  <summary>Click to see the full `main.go` file</summary>

```go
package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/crossplane/crossplane/apis/apiextensions/fn/io/v1alpha1"
	dummyv1alpha1 "github.com/upbound/provider-dummy/apis/iam/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/yaml"
)

var (
	Colors = []string{"red", "green", "blue", "yellow", "orange", "purple", "black", "white"}
)

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read stdin: %v", err)
		os.Exit(1)
	}
	obj := &v1alpha1.FunctionIO{}
	if err := yaml.Unmarshal(b, obj); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal stdin: %v", err)
		os.Exit(1)
	}
	colors := map[string]string{}
	for _, observed := range obj.Observed.Resources {
		r := &dummyv1alpha1.Robot{}
		if err := json.Unmarshal(observed.Resource.Raw, r); err != nil {
			fmt.Fprintf(os.Stderr, "failed to unmarshal observed resource: %v", err)
			os.Exit(1)
		}
		colors[observed.Name] = r.Spec.ForProvider.Color
	}
	for i, desired := range obj.Desired.Resources {
		r := &dummyv1alpha1.Robot{}
		if err := yaml.Unmarshal(desired.Resource.Raw, r); err != nil {
			fmt.Fprintf(os.Stderr, "failed to unmarshal desired resource: %v", err)
			os.Exit(1)
		}
		if colors[desired.Name] != "" {
			r.Spec.ForProvider.Color = colors[desired.Name]
		} else {
			r.Spec.ForProvider.Color = Colors[rand.Intn(len(Colors))]
		}
		// NOTE: We need to use a JSON marshaller here because runtiem.RawExtension
		// type expects a JSON blob.
		raw, err := json.Marshal(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal resource: %v", err)
			os.Exit(1)
		}
		obj.Desired.Resources[i].Resource = runtime.RawExtension{Raw: raw}
	}
	result, err := yaml.Marshal(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal resulting functionio: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(result))
}
```
</details>

```bash
# This is to make sure go.sum is tidied up after all the go get commands.
go mod tidy
```

Now let's make some changes in our `test.yaml` in order to validate that the
function does not manipulate the existing `Robot` objects that already have a
color set. We'll need to add an entry to `observed` section.

<details>
  <summary>Click to see the full `test.yaml` file</summary>

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
  resources:
    - name: one-robot
      resource:
        apiVersion: dummy.upbound.io/v1beta1
        kind: Robot
        spec:
          forProvider:
            color: yellow
        status:
          atProvider: {}
    - name: second-robot
      resource:
        apiVersion: dummy.upbound.io/v1beta1
        kind: Robot
        spec:
          forProvider:
            color: green
        status:
          atProvider: {}
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
    - name: second-robot
      resource:
        apiVersion: dummy.upbound.io/v1beta1
        kind: Robot
```
</details>

Let's run and validate that the first robot always has `yellow` and the other
one is changing with different runs.
```bash
cat test.yaml | go run main.go
```

## Try it Out

Let's build and push the function.
```bash
docker build --tag ${REGISTRY}/xfn-random:v0.1.0 .
docker push ${REGISTRY}/xfn-random:v0.1.0
```

Set the new image on our `Composition` object with `kubectl edit`. **Make sure to
use your DockerHub username instead of `${REGISTRY}`.**
```yaml
  ...
  functions:
    - name: my-random-function
      type: Container
      container:
        image: ${REGISTRY}/xfn-random:v0.1.0
```

Edit `Composition` to add a second `Robot` object but this time without its
color parameter set. The full `resources` array should look like the following:
```yaml
  resources:
    - name: one-robot
      base:
        apiVersion: iam.dummy.upbound.io/v1alpha1
        kind: Robot
        spec:
          forProvider:
            color: yellow
    - name: second-robot
      base:
        apiVersion: iam.dummy.upbound.io/v1alpha1
        kind: Robot
        spec:
          forProvider:
            color: ""
```

Let's create a new `RobotGroup` object and see what happens.
```bash
cat <<EOF | kubectl apply -f -
apiVersion: contribfest.crossplane.io/v1alpha1
kind: RobotGroup
metadata:
  name: my-robot-group
spec: {}
EOF
```

Let's list all `Robot`s and see what color each of them has.
```bash
kubectl get robots
```
```bash
kubectl get robots -o yaml
```

As you can see, one has `yellow` which was assigned in the `Composition` and the
other one has a random color assigned by our function and it persists.