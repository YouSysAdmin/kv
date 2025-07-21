package models

type Entity struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}
