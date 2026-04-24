package bus

type Listener interface {
	Handlers() map[string]Handler
}
