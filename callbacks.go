package registry

// ExpiryCallback is a function that is called when the node expires.
type ExpiryCallback func(*ExpiryNode) error

// NoOptExpriyCallback is a no opt callback.
func NoOptExpriyCallback(node *ExpiryNode) error {
	return nil
}
