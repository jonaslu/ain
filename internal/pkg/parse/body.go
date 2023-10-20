package parse

func (s *SectionedTemplate) getBody() []string {
	var body []string
	for _, bodySourceMarker := range *s.GetNamedSection(BodySection) {
		body = append(body, bodySourceMarker.LineContents)
	}

	return body
}
