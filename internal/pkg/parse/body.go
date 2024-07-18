package parse

func (s *sectionedTemplate) getBody() []string {
	var body []string
	for _, bodySourceMarker := range *s.getNamedSection(bodySection) {
		body = append(body, bodySourceMarker.lineContents)
	}

	return body
}
