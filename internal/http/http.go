package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"zadaniel0/internal/cache"

	"github.com/gin-gonic/gin"
)

type Http struct {
	cc *cache.Cache
	r  *gin.Engine
}

func NewHttp(cc *cache.Cache) *Http {
	return &Http{
		cc: cc,
		r:  gin.Default(),
	}
}

func (h *Http) Setup() {
	h.r.GET("/", h.RenderPage)
	h.r.GET("/find", h.RenderGetOrder)
	h.r.GET("/orders/:orderUID", h.HandleGetOrder)
}

func (h *Http) Run() error {
	if err := h.r.Run(); err != nil {
		return err
	}
	return nil
}

func (h *Http) HandleGetOrder(c *gin.Context) {
	orderUID := c.Param("orderUID")

	if orderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "orderUID is required",
		})
		return
	}

	order, err := h.cc.Get(c, orderUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "order not found",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Http) RenderPage(c *gin.Context) {
	orders := h.cc.GetAll()
	rows := make([]string, 0, len(orders))
	for _, order := range orders {
		bb, err := json.MarshalIndent(order, "", "  ")
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to marshal order")
			return
		}

		rows = append(rows, string(bb))
	}

	result := strings.Join(rows, "\n<hr>\n")

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
		<!DOCTYPE html>
		<html>
			<meta charset="UTF-8">
		  <body>
					<h1><a href="/">Домой!</a></h1>
					<form action="/find"><input type="text" placeholder="order_uid" name="order_uid"><button type="submit">Find</button></form>
					<pre>`+result+`</pre>
			</body>
		</html>
	`)
}

func (h *Http) RenderGetOrder(c *gin.Context) {
	orderUID := c.Query("order_uid")
	if len(orderUID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "order_uid is required",
		})
		return
	}

	order, err := h.cc.Get(c.Request.Context(), orderUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "order not found",
		})
		return
	}

	bb, err := json.MarshalIndent(order, "", "  ")
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to marshal order")
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, `
		<!DOCTYPE html>
		<html>
			<meta charset="UTF-8">
		  <body>
					<h1><a href="/">Домой!</a></h1>
					<form action="/find"><input type="text" placeholder="order_uid" name="order_uid"><button type="submit">Find</button></form>
					<pre>`+string(bb)+`</pre>
			</body>
		</html>
	`)
}
