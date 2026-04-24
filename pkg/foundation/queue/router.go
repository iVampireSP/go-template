package queue

type Router struct {
}

func newRouter() *Router {
	return &Router{}
}

func (r *Router) JobQueue(name string) string {
	return "foundation-default"
}
