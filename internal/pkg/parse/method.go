package parse

func (s *SectionedTemplate) getMethod() string {
	methodSourceMarkers := *s.GetNamedSection(MethodSection)

	if len(methodSourceMarkers) == 0 {
		return ""
	}

	if len(methodSourceMarkers) > 1 {
		s.SetFatalMessage("Found several lines under [Method]", methodSourceMarkers[0].SourceLineIndex)
		return ""
	}

	return methodSourceMarkers[0].LineContents
}
