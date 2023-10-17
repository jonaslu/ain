package main

func (s *SectionedTemplate) getBody() string {
	var body string
	for _, bodySourceMarker := range *s.GetNamedSection(BodySection) {
		body = body + bodySourceMarker.LineContents + "\n"
	}

	return body
}
