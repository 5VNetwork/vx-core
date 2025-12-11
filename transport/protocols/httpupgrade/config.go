package httpupgrade

func (c *HttpUpgradeConfig) GetNormalizedPath() string {
	path := c.Config.Path
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}
