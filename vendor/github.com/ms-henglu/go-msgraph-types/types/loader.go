package types

import (
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	typesEmbed "github.com/ms-henglu/go-msgraph-types/embed"
)

func DefaultMSGraphSchemaLoader() *MSGraphSchemaLoader {
	return &MSGraphSchemaLoader{
		staticFiles: typesEmbed.StaticFiles,
		mutex:       sync.Mutex{},
		cache:       make(map[*openapi3.Schema]*TypeBase),
	}
}

func NewMSGraphSchemaLoader(staticFiles embed.FS) *MSGraphSchemaLoader {
	return &MSGraphSchemaLoader{
		staticFiles: staticFiles,
		mutex:       sync.Mutex{},
		cache:       make(map[*openapi3.Schema]*TypeBase),
	}
}

type MSGraphSchemaLoader struct {
	schemaMap   map[string]*openapi3.T
	mutex       sync.Mutex
	staticFiles embed.FS
	cache       map[*openapi3.Schema]*TypeBase
}

func (r *MSGraphSchemaLoader) GetSchema(apiVersion string) *openapi3.T {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.schemaMap == nil {
		r.schemaMap = make(map[string]*openapi3.T)
	}
	if _, ok := r.schemaMap[apiVersion]; !ok {
		data, err := r.staticFiles.ReadFile(fmt.Sprintf("openapi/%s/openapi.yaml", apiVersion))
		if err != nil {
			log.Printf("[ERROR] failed to read schema: %+v", err)
			return nil
		}

		doc, err := openapi3.NewLoader().LoadFromData(data)
		if err != nil {
			log.Printf("[ERROR] failed to parse schema: %+v", err)
			return nil
		}
		r.schemaMap[apiVersion] = doc
	}
	return r.schemaMap[apiVersion]
}

func (r *MSGraphSchemaLoader) ListResources(apiVersion string) []ResourceType {
	schema := r.GetSchema(apiVersion)
	if schema == nil || schema.Paths == nil {
		return nil
	}

	var resources []ResourceType

	urlMap := make(map[string]string)
	m := make(map[string]*openapi3.PathItem)

	for path, pathItem := range schema.Paths.Map() {
		normalizedPath, _, _ := normalizeTemplatedPath(path)
		m[normalizedPath] = pathItem
		urlMap[normalizedPath] = path
	}

	for path, pathItem := range m {
		if pathItem.Post == nil {
			continue
		}

		if !strings.Contains(path, "/$ref") {
			itemPathItem := m[fmt.Sprintf("%s/%s", path, "{}")]
			if itemPathItem == nil {
				continue
			}

			if itemPathItem.Get == nil || itemPathItem.Delete == nil {
				continue
			}
		}

		resourceType := ResourceType{
			Type:        "resource",
			Url:         urlMap[path],
			Name:        pathItem.Post.Summary,
			Description: pathItem.Post.Description,
		}

		if pathItem.Post.ExternalDocs != nil {
			resourceType.ExternalDocs = &ExternalDocumentation{
				Description: pathItem.Post.ExternalDocs.Description,
				Url:         pathItem.Post.ExternalDocs.URL,
			}
		}

		resources = append(resources, resourceType)
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	return resources
}

func (r *MSGraphSchemaLoader) ListReadableResources(apiVersion string) []ResourceType {
	schema := r.GetSchema(apiVersion)
	if schema == nil || schema.Paths == nil {
		return nil
	}

	var resources []ResourceType

	for path, pathItem := range schema.Paths.Map() {
		if pathItem.Get == nil {
			continue
		}
		resourceType := ResourceType{
			Type:        "resource",
			Url:         path,
			Name:        pathItem.Get.Summary,
			Description: pathItem.Get.Description,
		}

		if pathItem.Get.ExternalDocs != nil {
			resourceType.ExternalDocs = &ExternalDocumentation{
				Description: pathItem.Get.ExternalDocs.Description,
				Url:         pathItem.Get.ExternalDocs.URL,
			}
		}

		resources = append(resources, resourceType)
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	return resources
}

func (r *MSGraphSchemaLoader) ListAPIVersions() []string {
	return []string{"v1.0", "beta"}
}

func (r *MSGraphSchemaLoader) GetResourceDefinition(apiVersion, url string) *ResourceType {
	schema := r.GetSchema(apiVersion)
	if schema == nil {
		return nil
	}

	postOperation := findOperation(schema, url, "POST")
	if postOperation == nil || postOperation.RequestBody == nil || postOperation.RequestBody.Value == nil || postOperation.RequestBody.Value.Content == nil {
		return nil
	}

	content := postOperation.RequestBody.Value.Content.Get("application/json")
	if content == nil || content.Schema == nil {
		return nil
	}

	requestBodyType := NewTypeBaseFromOpenAPISchema(content.Schema.Value, r.cache)
	if requestBodyType == nil {
		return nil
	}

	out := ResourceType{
		Type:        "resource",
		Url:         url,
		Name:        postOperation.Summary,
		Description: postOperation.Description,
		Body: &TypeReference{
			Type: *requestBodyType,
		},
	}

	if postOperation.ExternalDocs != nil {
		out.ExternalDocs = &ExternalDocumentation{
			Description: postOperation.ExternalDocs.Description,
			Url:         postOperation.ExternalDocs.URL,
		}
	}

	return &out
}

func findOperation(doc *openapi3.T, url string, method string) *openapi3.Operation {
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	pathItem := doc.Paths.Find(url)
	if pathItem == nil {
		return nil
	}

	var operation *openapi3.Operation
	switch method {
	case "GET":
		operation = pathItem.Get
	case "POST":
		operation = pathItem.Post
	case "PUT":
		operation = pathItem.Put
	case "DELETE":
		operation = pathItem.Delete
	case "PATCH":
		operation = pathItem.Patch
	case "OPTIONS":
		operation = pathItem.Options
	case "HEAD":
		operation = pathItem.Head
	default:
		return nil
	}

	return operation
}
