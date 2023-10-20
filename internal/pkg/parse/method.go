package parse

func (s *sectionedTemplate) getMethod() string {
	methodSourceMarkers := *s.getNamedSection(MethodSection)

	if len(methodSourceMarkers) == 0 {
		return ""
	}

	if len(methodSourceMarkers) > 1 {
		s.setFatalMessage("Found several lines under [Method]", methodSourceMarkers[0].SourceLineIndex)
		return ""
	}

	return methodSourceMarkers[0].LineContents
}
