package cmd

import "testing"

func TestDocMediaDownloadContextFlags(t *testing.T) {
	for _, name := range []string{"doc-token", "doc-type", "extra"} {
		if docMediaDownloadCmd.Flags().Lookup(name) == nil {
			t.Fatalf("doc media-download missing --%s flag", name)
		}
	}

	docType, err := docMediaDownloadCmd.Flags().GetString("doc-type")
	if err != nil {
		t.Fatalf("get doc-type flag: %v", err)
	}
	if docType != "docx" {
		t.Fatalf("doc-type default = %q, want docx", docType)
	}
}
