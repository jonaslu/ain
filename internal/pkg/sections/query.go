package main

func (s *SectionedTemplate) getQuery() []string {
	var query []string

	for _, querySourceMarker := range *s.GetNamedSection(QuerySection) {
		query = append(query, querySourceMarker.LineContents)
	}

	return query
}
