// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/jsonpath"
)

func TestParseSpecs(t *testing.T) {
	uu := map[string]struct {
		cols ColsSpecs
		err  error
		e    ColumnSpecs
	}{
		"empty": {
			e: ColumnSpecs{},
		},

		"plain": {
			cols: ColsSpecs{
				"a",
				"b",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"with-spec-plain": {
			cols: ColsSpecs{
				"a",
				"b:.metadata.name",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
					},
					Spec: "{.metadata.name}",
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"with-spec-fq": {
			cols: ColsSpecs{
				"a",
				"b:.metadata.name|NW",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
						Attrs: model1.Attrs{
							Wide:     true,
							Capacity: true,
							Align:    tview.AlignRight,
						},
					},
					Spec: "{.metadata.name}",
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"spec-type-no-wide": {
			cols: ColsSpecs{
				"a",
				"b:.metadata.name|T",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
						Attrs: model1.Attrs{
							Time: true,
						},
					},
					Spec: "{.metadata.name}",
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"plain-wide": {
			cols: ColsSpecs{
				"a",
				"b|W",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name:  "b",
						Attrs: model1.Attrs{Wide: true},
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"no-spec-kind-wide": {
			cols: ColsSpecs{
				"a",
				"b|NW",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "b",
						Attrs: model1.Attrs{
							Align:    tview.AlignRight,
							Capacity: true,
							Wide:     true,
						},
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},

		"toast-spec": {
			cols: ColsSpecs{
				"a",
				"b:{{crap.bozo}}|NW",
				"c",
			},
			err: errors.New(`unexpected path string, expected a 'name1.name2' or '.name1.name2' or '{name1.name2}' or '{.name1.name2}'`),
		},

		"no-spec": {
			cols: ColsSpecs{
				"a",
				"b|NW",
				"c",
			},
			e: ColumnSpecs{
				{
					Header: model1.HeaderColumn{
						Name: "a",
					},
				},
				{
					Header: model1.HeaderColumn{
						Name:  "b",
						Attrs: model1.Attrs{Align: tview.AlignRight, Capacity: true, Wide: true},
					},
				},
				{
					Header: model1.HeaderColumn{
						Name: "c",
					},
				},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cols, err := u.cols.parseSpecs()
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, cols)
		})
	}
}

func TestHydrateWithNilObject(t *testing.T) {
	uu := map[string]struct {
		o runtime.Object
		e string
	}{
		"nil-object": {
			o: nil,
			e: NAValue,
		},
		"nil-pointer": {
			o: (*testRuntimeObject)(nil),
			e: NAValue,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			parser := jsonpath.New("test").AllowMissingKeys(true)
			require.NoError(t, parser.Parse("{.metadata.name}"))

			cc := ColumnSpecs{
				{
					Header: model1.HeaderColumn{Name: "NAME"},
					Spec:   "{.metadata.name}",
				},
			}
			parsers := []*jsonpath.JSONPath{parser}
			rh := model1.Header{}
			row := &model1.Row{}

			cols, err := hydrate(u.o, cc, parsers, rh, row)
			require.NoError(t, err)
			require.Len(t, cols, 1)
			assert.Equal(t, u.e, cols[0].Value)
		})
	}
}

type testRuntimeObject struct{}

func (*testRuntimeObject) GetObjectKind() schema.ObjectKind {
	return nil
}

func (*testRuntimeObject) DeepCopyObject() runtime.Object {
	return nil
}
