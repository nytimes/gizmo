package nyttest

import (
	"github.com/NYTimes/gizmo/examples/nyt"
	"golang.org/x/net/context"
)

type Client struct {
	TestGetMostPopular        func(string, string, uint) ([]*nyt.MostPopularResult, error)
	TestSemanticConceptSearch func(string, string) ([]*nyt.SemanticConceptArticle, error)
}

func (t *Client) GetMostPopular(resourceType, section string, timeframe uint) ([]*nyt.MostPopularResult, error) {
	return t.TestGetMostPopular(resourceType, section, timeframe)
}

func (t *Client) SemanticConceptSearch(conceptType, concept string) ([]*nyt.SemanticConceptArticle, error) {
	return t.TestSemanticConceptSearch(conceptType, concept)
}

type CtxClient struct {
	TestGetMostPopular        func(string, string, uint) ([]*nyt.MostPopularResult, error)
	TestSemanticConceptSearch func(string, string) ([]*nyt.SemanticConceptArticle, error)
}

func (t *CtxClient) GetMostPopular(ctx context.Context, resourceType, section string, timeframe uint) ([]*nyt.MostPopularResult, error) {
	return t.TestGetMostPopular(resourceType, section, timeframe)
}

func (t *CtxClient) SemanticConceptSearch(ctx context.Context, conceptType, concept string) ([]*nyt.SemanticConceptArticle, error) {
	return t.TestSemanticConceptSearch(conceptType, concept)
}
