
package model

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDetectedObjects_Value(t *testing.T) {
	tests := []struct {
		name    string
		input   DetectedObjects
		wantErr bool
	}{
		{
			name: "successful conversion",
			input: DetectedObjects{
				{
					Label: "cat",
					Score: 0.98,
					Box: BoundingBox{
						Xmin: 10,
						Ymin: 20,
						Xmax: 100,
						Ymax: 120,
					},
				},
			},
		},
		{
			name:  "empty objects",
			input: DetectedObjects{},
		},
		{
			name:  "nil objects",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Value()

			if (err != nil) != tt.wantErr {
				t.Fatalf("Value() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			data, ok := got.([]byte)
			if !ok {
				t.Fatalf("Value() type = %T, want []byte", got)
			}

			var decoded DetectedObjects
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal Value() result: %v", err)
			}

			if !reflect.DeepEqual(decoded, tt.input) {
				t.Errorf("Value() result = %#v, want %#v", decoded, tt.input)
			}
		})
	}
}

func TestDetectedObjects_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    DetectedObjects
		wantErr bool
	}{
		{
			name: "successful scan",
			input: []byte(`[
				{
					"label": "cat",
					"score": 0.98,
					"box": {
						"xmin": 10,
						"ymin": 20,
						"xmax": 100,
						"ymax": 120
					}
				}
			]`),
			want: DetectedObjects{
				{
					Label: "cat",
					Score: 0.98,
					Box: BoundingBox{
						Xmin: 10,
						Ymin: 20,
						Xmax: 100,
						Ymax: 120,
					},
				},
			},
		},
		{
			name:    "empty objects",
			input:   []byte(`[]`),
			want:    DetectedObjects{},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`invalid json`),
			wantErr: true,
		},
		{
			name:    "unsupported input type",
			input:   "invalid type",
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DetectedObjects

			err := got.Scan(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scan() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
