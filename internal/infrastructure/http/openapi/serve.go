package openapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterOpenAPIRoutes(r gin.IRoutes) {
	r.GET("/openapi.yaml", func(c *gin.Context) {
		y := GetYAML()
		if len(y) == 0 {
			c.String(http.StatusInternalServerError, "openapi spec not available")
			return
		}
		c.Data(http.StatusOK, "application/yaml", y)
	})
	r.GET("/openapi.json", func(c *gin.Context) {
		j := GetJSON()
		if len(j) == 0 {
			c.String(http.StatusInternalServerError, "openapi spec not available")
			return
		}
		c.Data(http.StatusOK, "application/json", j)
	})

	// Minimal Swagger UI page using CDN, pointing to /openapi.json
	r.GET("/docs", func(c *gin.Context) {
		const html = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <title>API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
    <style>
      html, body { margin:0; padding:0; height:100%; }
      #swagger-ui { height: 100%; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({
        url: 'openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: 'BaseLayout'
      });
    </script>
  </body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})
}
