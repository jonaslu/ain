package parse

func (s *sectionedTemplate) getHeaders() []string {
	var headers []string

	for _, headerSourceMarker := range *s.getNamedSection(headersSection) {
		headers = append(headers, headerSourceMarker.lineContents)
	}

	return headers
}
