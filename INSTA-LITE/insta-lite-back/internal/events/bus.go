package events

// Bus générique : un channel bufferisé par type d'événement T.
type TypedBus[T any] struct {
	ch chan T
}

func NewTypedBus[T any](bufferSize int) *TypedBus[T] {
	return &TypedBus[T]{ch: make(chan T, bufferSize)}
}

// Publish bloque si le buffer est plein plutôt que de perdre l'événement
// silencieusement — même choix assumé que pour PostCreatedEvent.
func (b *TypedBus[T]) Publish(e T) {
	b.ch <- e
}

func (b *TypedBus[T]) Events() <-chan T {
	return b.ch
}
