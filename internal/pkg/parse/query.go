package parse

func (s *sectionedTemplate) getQuery() []string {
	var query []string

	for _, querySourceMarker := range *s.getNamedSection(querySection) {
		query = append(query, querySourceMarker.lineContents)
	}

	return query
}
