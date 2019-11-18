package roaring

type Container interface {
	Add(container Container) Container

	getContainer() container
}

type containerWrapper struct {
	container container
}

func (c *containerWrapper) Add(container Container) Container {
	result := c.container.iand(container.getContainer())
	return &containerWrapper{container: result}
}

func (c *containerWrapper) getContainer() container {
	return c.container
}
