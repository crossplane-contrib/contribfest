# Building a Function that Creates N Resources

In `Composition`, the list of resources is always static meaning that you cannot
have a conditional logic that creates a different number of resources based on
the user input. For example, there can be a parameter to make a database public
or private and depending on the choice, you may need to create different number
of firewall rules.

In this example, we will build a function that will get a count from the user
and create that many managed resources.

We will use DockerHub to push & pull the images we build. The following environment
variable will be required to make sure the commands work with your own images.
```bash
export REGISTRY=<your-dockerhub-username>
```

Let's build on top of our no-op function.
```bash
cp -a xfn-random xfn-many
cd xfn-many
```

Change the function name to `xfn-many` in all files.
```bash
# On Mac
sed -i '' 's/xfn-random/xfn-many/g' *
# On Linux
sed -i 's/xfn-random/xfn-many/g' *
```

NOTE: The rest of the guide assumes that you already have the `CompositeResourceDefinition`
created from the previous tutorials. If you don't, you can
go back [installation section](02-xfn-noop.md#installation) and create it.

### Creating New Resources

In the earlier example, we parsed existing `Robot` objects and made changes on
the desired ones that were sent to us. In this one, we will add new resources
to the list of desired resources that were not sent to us by the `resources`
defined in `Composition`.

We are continuing from where we left off in the previous tutorial. The following
is the `main.go` file that we had.
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

Different from the previous tutorials, we will now parse the composite resource
since we'll need the count from the user. The following code snippet uses
functionality from Crossplane Runtime and Kubernetes API Machinery to parse
an arbitrary resource because differently from `Robot`, composite resources do
not have typed Go structs - they are defined as `CustomResourceDefinition`s in
the API server.
```go
	robotGroup := composite.New()
	if err := yaml.Unmarshal(obj.Observed.Composite.Resource.Raw, &robotGroup.Unstructured); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal observed composite: %v", err)
		os.Exit(1)
	}
	count, err := fieldpath.Pave(robotGroup.Object).GetInteger("spec.count")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get the count from observed composite: %v", err)
		os.Exit(1)
	}
```

We will now create `count` number of `Robot` resources and add them to the
desired list. The key point here is to make sure we are adding to the observed
`Robot`s so that we don't repeatedly create new `Robot`s and delete the old ones.

Let's calculate how many new `Robot`s we need to create.
```go
	var robots []v1alpha1.DesiredResource
	for _, r := range obj.Observed.Resources {
		robots = append(robots, v1alpha1.DesiredResource{
			Name:     r.Name,
			Resource: r.Resource,
		})
	}
	add := int(count) - len(robots)
```

Every `Robot` entry in the desired list should have its own identifier, so let's
write a small function to generate a random suffix.
```bash
go get github.com/crossplane/crossplane-runtime/pkg/password
go mod tidy
```
```go
func generateSuffix() (string, error) {
    return password.Settings{
        CharacterSet: "abcdefghijklmnopqrstuvwxyz0123456789",
        Length:       5,
    }.Generate()
}
```

Now, let's create new `Robot`s and add them to the desired list.
```go
	for i := 0; i < add; i++ {
		suf, err := generateSuffix()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate random suffix for name: %v", err)
			os.Exit(1)
		}
		r := &dummyv1alpha1.Robot{
			Spec: dummyv1alpha1.RobotSpec{
				ForProvider: dummyv1alpha1.RobotParameters{
					Color: Colors[rand.Intn(len(Colors))],
				},
			},
		}
        r.SetName(robotGroup.GetName() + "-" + suf)
        r.SetGroupVersionKind(dummyv1alpha1.RobotGroupVersionKind)
		// NOTE: We need to use a JSON marshaller here because runtiem.RawExtension
		// type expects a JSON blob.
		raw, err := json.Marshal(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal resource: %v", err)
			os.Exit(1)
		}
		robots = append(robots, v1alpha1.DesiredResource{
			Name: "robot-" + suf,
			Resource: runtime.RawExtension{
				Raw: raw,
			},
		})
	}
	obj.Desired.Resources = robots
```

The following is the full `main.go` file to make sure you got all of it right.

<details>
  <summary>Click to see the full `main.go` file</summary>

```go
package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/resource/unstructured/composite"
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
	robotGroup := composite.New()
	if err := yaml.Unmarshal(obj.Observed.Composite.Resource.Raw, &robotGroup.Unstructured); err != nil {
		fmt.Fprintf(os.Stderr, "failed to unmarshal observed composite: %v", err)
		os.Exit(1)
	}
	count, err := fieldpath.Pave(robotGroup.Object).GetInteger("spec.count")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get the count from observed composite: %v", err)
		os.Exit(1)
	}
	var robots []v1alpha1.DesiredResource
	for _, r := range obj.Observed.Resources {
		robots = append(robots, v1alpha1.DesiredResource{
			Name:     r.Name,
			Resource: r.Resource,
		})
	}
	add := int(count) - len(robots)
	for i := 0; i < add; i++ {
		r := &dummyv1alpha1.Robot{
			Spec: dummyv1alpha1.RobotSpec{
				ForProvider: dummyv1alpha1.RobotParameters{
					Color: Colors[rand.Intn(len(Colors))],
				},
			},
		}
		r.SetName(robotGroup.GetName() + "-" + suf)
		r.SetGroupVersionKind(dummyv1alpha1.RobotGroupVersionKind)
		// NOTE: We need to use a JSON marshaller here because runtiem.RawExtension
		// type expects a JSON blob.
		raw, err := json.Marshal(r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal resource: %v", err)
			os.Exit(1)
		}
		suf, err := generateSuffix()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to generate random suffix for name: %v", err)
			os.Exit(1)
		}
		robots = append(robots, v1alpha1.DesiredResource{
			Name: "robot-" + suf,
			Resource: runtime.RawExtension{
				Raw: raw,
			},
		})
	}
	obj.Desired.Resources = robots
	result, err := yaml.Marshal(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal resulting functionio: %v", err)
		os.Exit(1)
	}
	fmt.Print(string(result))
}

func generateSuffix() (string, error) {
	return password.Settings{
		CharacterSet: "abcdefghijklmnopqrstuvwxyz0123456789",
		Length:       5,
	}.Generate()
}
```
</details>

We will make some modifications to our `test.yaml` file to make sure `count` is
given in the composite resource. In the test input below, we see that there is
already a single `Robot` but the user requested to create `5` robots. So, we expect
our function to add the existing robot `4` more new ones under desired resources.

Note that `desired.resources` is empty because the standard `Composition` does
not send any desired robots since the count may be `0`.

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
      spec:
        parameters:
          count: 5
  resources:
    - name: robot-8sd7h
      resource:
        apiVersion: dummy.upbound.io/v1beta1
        kind: Robot
        spec:
          forProvider:
            color: yellow
        status:
          atProvider: {}
desired:
  composite:
    resource:
      apiVersion: contribfest.crossplane.io/v1alpha1
      kind: XRobotGroup
      metadata:
        name: somename
      spec:
        parameters:
          count: 5
```

Let's run our function and see the result.

```bash
cat test.yaml | go run main.go
```

## Try it Out

Let's build and push the function.
```bash
docker build --tag ${REGISTRY}/xfn-many:v0.1.0 .
docker push ${REGISTRY}/xfn-many:v0.1.0
```

Let's create a new `Composition` that uses our function. Make sure to delete
the existing one first.
```bash
kubectl delete composition robotgroup
```
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
  functions:
  - name: my-many-function
    type: Container
    container:
      image: ${REGISTRY}/xfn-many:v0.1.0
EOF
```

Let's create a new `RobotGroup` object and see what happens.
```bash
cat <<EOF | kubectl apply -f -
apiVersion: contribfest.crossplane.io/v1alpha1
kind: RobotGroup
metadata:
  name: my-robot-group
spec:
  count: 5
EOF
```

Let's list all `Robot`s and see how many we'll see. There should be exactly `5`.
```bash
kubectl get robots
```

### Cleanup

Delete the `RobotGroup` instance.
```bash
kubectl delete robotgroup my-robot-group
```