package nyt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type (
	Client interface {
		GetMostPopular(string, string, uint) ([]*MostPopularResult, error)
		SemanticConceptSearch(string, string) ([]*SemanticConceptArticle, error)
	}
	ClientImpl struct {
		mostPopularToken string
		semanticToken    string
	}
)

func NewClient(mostPopToken, semanticToken string) Client {
	return &ClientImpl{mostPopToken, semanticToken}
}

func (c *ClientImpl) GetMostPopular(resourceType string, section string, timePeriodDays uint) ([]*MostPopularResult, error) {
	var (
		res MostPopularResponse
	)
	uri := fmt.Sprintf("/svc/mostpopular/v2/%s/%s/%d.json?api-key=%s",
		resourceType,
		section,
		timePeriodDays,
		c.mostPopularToken)

	rawRes, err := c.do(uri)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawRes, &res)
	return res.Results, err
}

func (c *ClientImpl) SemanticConceptSearch(conceptType, concept string) ([]*SemanticConceptArticle, error) {
	var (
		res SemanticConceptResponse
	)
	uri := fmt.Sprintf("/svc/semantic/v2/concept/name/nytd_%s/%s.json?fields=article_list&api-key=%s",
		conceptType,
		concept,
		c.semanticToken)

	rawRes, err := c.do(uri)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(rawRes, &res)
	if err != nil {
		return nil, err
	}
	if len(res.Results) == 0 {
		return []*SemanticConceptArticle{}, nil
	}

	return res.Results[0].ArticleList.Results, nil
}

func (c *ClientImpl) do(uri string) (body []byte, err error) {
	hc := http.Client{
		Timeout: 5 * time.Second,
	}
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
