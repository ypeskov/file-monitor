package render

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		log.Errorf("Error rendering component: %e\n", err)
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}
