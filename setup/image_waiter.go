package setup

// ImageWaiter can wait for an image to complete.
type ImageWaiter interface {
	WaitForImage(ref string) error
}
