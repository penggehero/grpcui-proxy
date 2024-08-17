package main

import (
	"flag"
	"html/template"
	"net/http"
	"strconv"

	"github.com/fullstorydev/grpcui"
	"github.com/fullstorydev/grpcui/standalone"
	"github.com/gin-gonic/gin"
)

var port = flag.Int("port", 8080, "The server port")

func main() {
	flag.Parse()
	engine := gin.Default()
	// load html and static files
	engine.LoadHTMLGlob("templates/*")
	engine.Static("/static", "static")

	engine.GET("/", indexHandler)
	engine.POST("/invoke/:method", invokeHandler)
	engine.GET("/metadata", metadataHandler)
	engine.GET("/grpcui", grpcuiHandler)
	engine.GET("/examples", examplesHandler)

	err := engine.Run(":" + strconv.Itoa(*port))
	if err != nil {
		panic(err)
	}
}

// index is the handler for the index page
func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// examples returns a list of examples
func examplesHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, "[]")
}

// grpcuiHandler is the handler for the grpcui page
func grpcuiHandler(c *gin.Context) {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": "endpoint is required"})
		return
	}
	option, err := NewGrpcuiProxyOption(c, endpoint)
	if err != nil {
		c.HTML(http.StatusBadRequest, "index.html", gin.H{"error": err.Error()})
		return
	}
	defer option.cc.Close()
	webFormHTML := grpcui.WebFormContentsWithOptions("invoke", "metadata", endpoint, option.methods,
		grpcui.WebFormOptions{})
	data := standalone.WebFormContainerTemplateData{
		Target:          endpoint,
		WebFormContents: template.HTML(webFormHTML),
		AddlResources:   []template.HTML{},
	}
	c.HTML(http.StatusOK, "index-template.html", gin.H{
		"WebFormContents": data.WebFormContents,
		"Target":          data.Target,
		"AddlResources":   data.AddlResources,
	})

}

// invokeHandler is the handler for the invoke rpc call
func invokeHandler(c *gin.Context) {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		c.HTML(http.StatusBadRequest, "index-template.html", gin.H{"error": "endpoint is required"})
		return
	}
	option, err := NewGrpcuiProxyOption(c, endpoint)
	if err != nil {
		c.HTML(http.StatusBadRequest, "index-template.html", gin.H{"error": err.Error()})
		return
	}
	defer option.cc.Close()
	rpcInvokeHandler := http.StripPrefix("/invoke",
		grpcui.RPCInvokeHandlerWithOptions(option.cc, option.methods, grpcui.InvokeOptions{
			EmitDefaults: true,
		}))
	rpcInvokeHandler.ServeHTTP(c.Writer, c.Request)
}

// metadataHandler is the handler for the metadata page
func metadataHandler(c *gin.Context) {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		c.HTML(http.StatusBadRequest, "index-template.html", gin.H{"error": "endpoint is required"})
		return
	}
	option, err := NewGrpcuiProxyOption(c, endpoint)
	if err != nil {
		c.HTML(http.StatusBadRequest, "index-template.html", gin.H{"error": err.Error()})
		return
	}
	defer option.cc.Close()
	rpcMetadataHandler := grpcui.RPCMetadataHandler(option.methods, option.files)
	rpcMetadataHandler.ServeHTTP(c.Writer, c.Request)
}
