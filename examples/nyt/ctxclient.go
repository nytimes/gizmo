package nyt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type (
	ContextClient interface {
		GetMostPopular(context.Context, string, string, uint) ([]*MostPopularResult, error)
		SemanticConceptSearch(context.Context, string, string) ([]*SemanticConceptArticle, error)
	}
	ContextClientImpl struct {
		mostPopularToken string
		semanticToken    string
	}
)

func NewContextClient(mostPopToken, semanticToken string) ContextClient {
	return &ContextClientImpl{mostPopToken, semanticToken}
}

func (c *ContextClientImpl) GetMostPopular(ctx context.Context, resourceType string, section string, timePeriodDays uint) ([]*MostPopularResult, error) {
	var (
		res MostPopularResponse
	)
	uri := fmt.Sprintf("/svc/mostpopular/v2/%s/%s/%d.json?api-key=%s",
		resourceType,
		section,
		timePeriodDays,
		c.mostPopularToken)

	rawRes, err := c.do(ctx, uri)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawRes, &res)
	return res.Results, err
}

func (c *ContextClientImpl) SemanticConceptSearch(ctx context.Context, conceptType, concept string) ([]*SemanticConceptArticle, error) {
	var (
		res SemanticConceptResponse
	)
	uri := fmt.Sprintf("/svc/semantic/v2/concept/name/nytd_%s/%s.json?fields=article_list&api-key=%s",
		conceptType,
		concept,
		c.semanticToken)

	rawRes, err := c.do(ctx, uri)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawRes, &res)
	if len(res.Results) == 0 {
		log.Debugf(ctx, "%s", err)
		return nil, errors.New("no results")
	}

	return res.Results[0].ArticleList.Results, nil
}

func (c *ContextClientImpl) do(ctx context.Context, uri string) (body []byte, err error) {
	hc := urlfetch.Client(ctx)
	req, err := http.NewRequest("GET", "https://api.nytimes.com"+uri, nil)
	if err != nil {
		return nil, err
	}

	var res *http.Response
	res, err = hc.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := res.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return ioutil.ReadAll(res.Body)
}
