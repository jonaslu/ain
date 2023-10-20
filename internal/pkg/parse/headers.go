package parse

func (s *SectionedTemplate) getHeaders() []string {
	var headers []string

	for _, headerSourceMarker := range *s.GetNamedSection(HeadersSection) {
		headers = append(headers, headerSourceMarker.LineContents)
	}

	return headers
}
