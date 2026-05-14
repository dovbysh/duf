package is_medical

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed z_LabAnalysisResults.md
var labAnalysisResults []byte

//go:embed z_Protocol.md
var protocol []byte

//go:embed z_otherMedical.md
var otherMedical []byte

func TestGetDocumentClassification(t *testing.T) {
	type str struct {
		res  string
		want DocumentType
	}
	tt := []str{
		{
			res:  string(labAnalysisResults),
			want: DocLabAnalysisResults,
		},
		{
			res:  string(protocol),
			want: DocProtocol,
		},
		{
			res:  string(otherMedical),
			want: DocOtherMedical,
		},
	}
	for _, s := range tt {
		t.Run(string(s.want), func(t *testing.T) {
			c, err := GetDocumentClassification(s.res)
			require.NoError(t, err)
			require.NotNil(t, c)
			require.NotNil(t, c.DocumentType)
			require.Equal(t, s.want, *c.DocumentType)
		})
	}
}
