package v1alpha3

// Hub marks v1alpha3 as the conversion hub for NetworkNamespace. Every other
// served version converts to and from this one.
func (*NetworkNamespace) Hub() {}
