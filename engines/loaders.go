package engines

import (
	"context"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/zefrenchwan/scrutateur.git/storage"
)

// MapUrlToPath transforms an url asking for resource to a local path.
// PARAMETERS:
// urlBase is the path base to delete from the url
// localBase is the local path base to replace within the url
// url is the resource asked by user
// RESULT:
// the local path of the resource if any
// an error if said file does not exist, or if url is suspicious (like an attempt to go through the local FS)
func MapUrlToPath(urlBase, localBase, url string) (string, error) {
	if strings.Contains(url, "//") {
		return "", errors.New("suspicious url")
	} else if strings.Contains(url, "..") {
		return "", errors.New("suspicious url")
	} else if !strings.HasPrefix(url, "/") {
		return "", errors.New("absolute url only")
	}

	var siteUrlBase string
	if !strings.HasSuffix(urlBase, "/") {
		siteUrlBase = urlBase + "/"
	} else {
		siteUrlBase = urlBase
	}

	var localUrlBase string
	if !strings.HasSuffix(localBase, "/") {
		localUrlBase = localBase + "/"
	} else {
		localUrlBase = localBase
	}

	if len(siteUrlBase) >= len(url) || !strings.HasPrefix(url, siteUrlBase) {
		return "", errors.New("url is out of scope")
	} else {
		rawPath := path.Join(localUrlBase, url[len(siteUrlBase):])
		if _, err := os.Stat(rawPath); err != nil {
			return "", err
		} else if !strings.Contains(path.Dir(rawPath), localBase) {
			return "", errors.New("out of static directory")
		}

		return rawPath, nil
	}
}

// ResourceLoader loads a resource
type ResourceLoader interface {
	// Load loads the resource if accepted
	Load(url string) ([]byte, error)
	// Accept returns whether that loader would be able to load the resource
	Accept(url string) bool
}

// CompositeLoader loads content from the first child that is able to
type CompositeLoader struct {
	// loaders candidates
	loaders []ResourceLoader
}

// NewCompositeLoader builds a composite loader from existing loaders
func NewCompositeLoader(loaders ...ResourceLoader) CompositeLoader {
	if len(loaders) == 0 {
		panic(errors.New("no loader in composite loader"))
	}

	var result CompositeLoader
	result.loaders = append(result.loaders, loaders...)
	return result
}

// Load gets content from the first loader that may load the content
func (c *CompositeLoader) Load(url string) ([]byte, error) {
	for _, loader := range c.loaders {
		if loader.Accept(url) {
			return loader.Load(url)
		}
	}

	return nil, errors.New("no matching loader")
}

// Accept returns true if one loader at least may load content
func (c *CompositeLoader) Accept(url string) bool {
	for _, loader := range c.loaders {
		if loader.Accept(url) {
			return true
		}
	}

	return false
}

// LocalResourceLoader gets content from local storage
type LocalResourceLoader struct {
	// urlBase defines where to start the resource from (for instance "/resources/")
	urlBase string
	// localBase defines where to start looking on local storage (for instance "/app/static/")
	localBase string
}

func NewLocalResourceLoader(urlBase, localBase string) ResourceLoader {
	return LocalResourceLoader{urlBase: urlBase, localBase: localBase}
}

// mapToLocalPath returns the local path to look for a file, or error for access failure
// (or some moron trying to get out from the static directory)
func (rl LocalResourceLoader) mapToLocalPath(url string) (string, error) {
	return MapUrlToPath(rl.urlBase, rl.localBase, url)
}

// Load just gets content from file by name
func (rl LocalResourceLoader) Load(url string) ([]byte, error) {
	if localPath, err := rl.mapToLocalPath(url); err != nil {
		return nil, err
	} else {
		return os.ReadFile(localPath)
	}
}

// Accept returns true if the resource is on the storage system
func (rl LocalResourceLoader) Accept(url string) bool {
	_, err := rl.mapToLocalPath(url)
	return err == nil
}

// CachedResourceLoader gets data from a cache
type CachedResourceLoader struct {
	Cache storage.CacheStorage
}

// Load gets a file from a cache
func (c CachedResourceLoader) Load(url string) ([]byte, error) {
	return c.Cache.GetValue(context.Background(), url)
}

// Accept returns true if that url is stored in the cache
func (c CachedResourceLoader) Accept(url string) bool {
	return c.Cache.Has(context.Background(), url)
}
