package parse

func (s *sectionedTemplate) getHost() string {
	var host string
	for _, hostSourceMarker := range *s.getNamedSection(HostSection) {
		host = host + hostSourceMarker.LineContents
	}

	return host
}
