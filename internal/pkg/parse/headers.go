package parse

func (s *sectionedTemplate) getHeaders() []string {
	var headers []string

	for _, headerSourceMarker := range *s.getNamedSection(HeadersSection) {
		headers = append(headers, headerSourceMarker.LineContents)
	}

	return headers
}
