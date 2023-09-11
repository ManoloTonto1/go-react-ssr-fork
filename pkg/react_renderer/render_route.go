package react_renderer

import (
	"encoding/json"
	"fmt"
	"gossr/config"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Config struct {
    File string
    MetaTags map[string]string
    Props interface{}
}

// Uses esbuild to render the React page at the given file path and inserts it into an html page
func RenderRoute(c *gin.Context, renderConfig Config) {
    props := "null"
    if renderConfig.Props != nil {
        propsJSON, err := json.Marshal(renderConfig.Props)
        if err != nil {
            c.JSON(500, gin.H{"error": "Invalid prop types"})
            return
        }
        props = string(propsJSON)
    }
    // Get the full path of the file
    filePath := getFullFilePath(config.Config.Web.SrcDirectory + "/" + renderConfig.File)
    updateRouteToFileMap(c.Request.URL.Path, filePath)
    cachedBuild, ok := checkForCachedBuild(filePath)
    if !ok {
        var err error
        var metafile string
        cachedBuild, metafile, err = BuildFile(filePath, props)
        if err != nil {
            c.HTML(500, "index.html", gin.H{
                "src": template.JS(fmt.Sprintf("showError(`%s`)", err.Error())),
            })
            return
        }
        updateFileToDependenciesMap(filePath, getDependenciesFromMetafile(metafile))
        cacheBuild(filePath, cachedBuild)
    }
    title := getTitle(renderConfig.MetaTags)
    delete(renderConfig.MetaTags, "title")
    c.HTML(http.StatusOK, "index.html", gin.H{
        "route": c.Request.URL.Path,
        "title": title,
        "metaTags": getMetaTags(renderConfig.MetaTags),
        "ogMetaTags": getOGMetaTags(renderConfig.MetaTags),
        "src": template.JS(cachedBuild.CompiledJS),
        "css": template.CSS(cachedBuild.CompiledCSS),
    })
}

func getTitle(metaTags map[string]string) string {
    title, ok := metaTags["title"]
    if ok {
        return title
    }
    return "Go SSR App"
}

func getMetaTags(metaTags map[string]string) map[string]string {
    newMetaTags := make(map[string]string)
    for key, value := range metaTags {
        if !strings.HasPrefix(key, "og:") {
            newMetaTags[key] = value
        }
    }
    return newMetaTags
}

func getOGMetaTags(metaTags map[string]string) map[string]string {
    newMetaTags := make(map[string]string)
    for key, value := range metaTags {
        if strings.HasPrefix(key, "og:") {
            newMetaTags[key] = value
        }
    }
    return newMetaTags
}