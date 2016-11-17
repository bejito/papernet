package bleve

import (
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search/query"

	"github.com/bobinette/papernet"
)

type PaperSearch struct {
	Repository papernet.PaperRepository

	index bleve.Index
}

func (s *PaperSearch) Open(path string) error {
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		indexMapping := createMapping()
		index, err = bleve.New(path, indexMapping)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	s.index = index
	return nil
}

func (s *PaperSearch) Close() error {
	if s.index == nil {
		return nil
	}

	return s.index.Close()
}

func (s *PaperSearch) Index(paper *papernet.Paper) error {
	data := map[string]interface{}{
		"title": paper.Title,
	}

	return s.index.Index(strconv.Itoa(paper.ID), data)
}

func (s *PaperSearch) Search(titlePrefix string) ([]int, error) {
	var q query.Query
	if titlePrefix != "" {
		tokens := splitNonEmpty(titlePrefix, " ")
		conjuncts := make([]query.Query, len(tokens))
		for i, token := range tokens {
			conjuncts[i] = &query.PrefixQuery{
				Prefix: token,
				Field:  "title",
			}
		}
		q = query.NewConjunctionQuery(conjuncts)
	} else {
		q = query.NewMatchAllQuery()
	}

	search := bleve.NewSearchRequest(q)
	search.SortBy([]string{"id"})
	searchResults, err := s.index.Search(search)
	if err != nil {
		return nil, err
	}

	ids := make([]int, searchResults.Total)
	for i, hit := range searchResults.Hits {
		ids[i], err = strconv.Atoi(hit.ID)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// ------------------------------------------------------------------------------------------------
// Mapping
// ------------------------------------------------------------------------------------------------

func createMapping() mapping.IndexMapping {
	// a generic reusable mapping for english text -- from blevesearch/beer-search
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	// Paper mapping
	paperMapping := bleve.NewDocumentMapping()
	paperMapping.AddFieldMappingsAt("title", englishTextFieldMapping)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultMapping = paperMapping
	return indexMapping
}

// ------------------------------------------------------------------------------------------------
// Helpers
// ------------------------------------------------------------------------------------------------

func splitNonEmpty(s string, sep string) []string {
	splitted := strings.Split(s, sep)
	res := make([]string, 0, len(splitted))
	for _, str := range splitted {
		if str == "" {
			continue
		}

		res = append(res, str)
	}
	return res
}
