package api

func (h *Handler) setRoutes() error {
	public := h.server.Group("/api")

	public.POST("/login", h.auth.PostLogin)
	public.POST("/register", h.auth.PostRegister)

	return nil
}
