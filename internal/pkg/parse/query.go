package parse

func (s *sectionedTemplate) getQuery() []string {
	var query []string

	for _, querySourceMarker := range *s.getNamedSection(QuerySection) {
		query = append(query, querySourceMarker.LineContents)
	}

	return query
}
