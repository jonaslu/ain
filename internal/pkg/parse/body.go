package parse

func (s *sectionedTemplate) getBody() []string {
	var body []string
	for _, bodySourceMarker := range *s.getNamedSection(BodySection) {
		body = append(body, bodySourceMarker.LineContents)
	}

	return body
}
