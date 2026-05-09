package types

type KV[V any] struct {
	Key   string
	Value V
}
