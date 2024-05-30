package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type OpenAPI struct {
	Swagger string `json:"swagger"`
	Info    struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"info"`
	Paths map[string]interface{} `json:"paths"`
}

func fetchOpenAPIV2JSON(clientset *kubernetes.Clientset) ([]byte, error) {
	req := clientset.Discovery().RESTClient().Get().AbsPath("/openapi/v2")
	return req.Do(context.TODO()).Raw()
}

func parseOpenAPI(data []byte) (*OpenAPI, error) {
	var openAPI OpenAPI
	err := json.Unmarshal(data, &openAPI)
	if err != nil {
		return nil, err
	}
	return &openAPI, nil
}

func main() {
	r := gin.Default()

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %s\n", err.Error())
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %s\n", err.Error())
		return
	}

	r.Any("/*path", func(c *gin.Context) {
		forwardRequest(c, clientset)
	})

	r.GET("/swagger.json", func(c *gin.Context) {
		data, err := fetchOpenAPIV2JSON(clientset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		openAPI, err := parseOpenAPI(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, openAPI)
	})

	// Integrate Swagger UI
	// r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger.json")))

	log.Fatal(r.Run(":8080"))
}

func forwardRequest(c *gin.Context, clientset *kubernetes.Clientset) {
	req := clientset.Discovery().RESTClient().Verb(c.Request.Method).AbsPath(c.Request.RequestURI).Body(c.Request.Body).SetHeader("Content-Type", c.ContentType())

	result := req.Do(context.TODO())
	rawBody, err := result.Raw()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", rawBody)
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
