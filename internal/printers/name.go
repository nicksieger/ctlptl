/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package printers

import (
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/tilt-dev/ctlptl/pkg/api"
)

// NOTE(nick): A fork of the NamePrinter that supports ctlptl types, which don't have a full object metadata

type Named interface {
	GetName() string
}

var _ Named = &api.Cluster{}
var _ Named = &api.Registry{}

// NamePrinter is an implementation of ResourcePrinter which outputs "resource/name" pair of an object.
type NamePrinter struct {
	// ShortOutput indicates whether an operation should be
	// printed along side the "resource/name" pair for an object.
	ShortOutput bool
	// Operation describes the name of the action that
	// took place on an object, to be included in the
	// finalized "successful" message.
	Operation string
}

// PrintObj is an implementation of ResourcePrinter.PrintObj which decodes the object
// and print "resource/name" pair. If the object is a List, print all items in it.
func (p *NamePrinter) PrintObj(obj runtime.Object, w io.Writer) error {
	if obj.GetObjectKind().GroupVersionKind().Empty() {
		return fmt.Errorf("missing apiVersion or kind; try GetObjectKind().SetGroupVersionKind() if you know the type")
	}

	name := "<unknown>"
	if named, ok := obj.(Named); ok {
		if n := named.GetName(); len(n) > 0 {
			name = n
		}
	}
	if acc, err := meta.Accessor(obj); err == nil {
		if n := acc.GetName(); len(n) > 0 {
			name = n
		}
	}

	return printObj(w, name, p.Operation, p.ShortOutput, GetObjectGroupKind(obj))
}

func GetObjectGroupKind(obj runtime.Object) schema.GroupKind {
	if obj == nil {
		return schema.GroupKind{Kind: "<unknown>"}
	}
	groupVersionKind := obj.GetObjectKind().GroupVersionKind()
	if len(groupVersionKind.Kind) > 0 {
		return groupVersionKind.GroupKind()
	}

	if uns, ok := obj.(*unstructured.Unstructured); ok {
		if len(uns.GroupVersionKind().Kind) > 0 {
			return uns.GroupVersionKind().GroupKind()
		}
	}

	return schema.GroupKind{Kind: "<unknown>"}
}

func printObj(w io.Writer, name string, operation string, shortOutput bool, groupKind schema.GroupKind) error {
	if len(groupKind.Kind) == 0 {
		return fmt.Errorf("missing kind for resource with name %v", name)
	}

	if len(operation) > 0 {
		operation = " " + operation
	}

	if shortOutput {
		operation = ""
	}

	if len(groupKind.Group) == 0 {
		fmt.Fprintf(w, "%s/%s%s\n", strings.ToLower(groupKind.Kind), name, operation)
		return nil
	}

	fmt.Fprintf(w, "%s.%s/%s%s\n", strings.ToLower(groupKind.Kind), groupKind.Group, name, operation)
	return nil
}
