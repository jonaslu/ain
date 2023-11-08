package parse

func (s *sectionedTemplate) getHost() string {
	var host string
	for _, hostSourceMarker := range *s.getNamedSection(hostSection) {
		host = host + hostSourceMarker.LineContents
	}

	return host
}
