package main

func (s *SectionedTemplate) getHost() string {
	var host string
	for _, hostSourceMarker := range *s.GetNamedSection(HostSection) {
		host = host + hostSourceMarker.LineContents
	}

	return host
}
