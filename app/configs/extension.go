package configs

func (h *HandlerConfig) GetTag() string {
	if h.GetOutbound() != nil {
		return h.GetOutbound().Tag
	} else if h.GetChain() != nil {
		return h.GetChain().Tag
	}
	return ""
}
