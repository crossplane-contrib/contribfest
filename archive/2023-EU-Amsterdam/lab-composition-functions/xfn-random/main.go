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
