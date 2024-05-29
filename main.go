package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io/ioutil"
	"log"
	"net/http"
)

var swaggerJSON []byte

func fetchSwaggerJSON(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func main() {
	r := gin.Default()

	var err error
	swaggerJSON, err = fetchSwaggerJSON("http://localhost:8000/openapi/v2")
	if err != nil {
		log.Fatalf("Failed to fetch Swagger JSON: %v", err)
	}

	// Serve the Swagger JSON
	r.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", swaggerJSON)
	})

	// Integrate Swagger UI
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger")))
	r.Use(requestForwarderMiddleware("http://localhost:8000"))
	log.Fatal(r.Run(":8080"))
}

func requestForwarderMiddleware(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new request based on the original request
		req, err := http.NewRequest(c.Request.Method, targetURL+c.Request.RequestURI, c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy headers from the original request
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
			return
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// Set the response status
		c.Status(resp.StatusCode)

		// Set the response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Set the response body
		c.Writer.Write(body)

		// Abort the context to prevent further processing
		c.Abort()
	}
}
