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
