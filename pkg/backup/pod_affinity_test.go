package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGetNodeAffinityForBackupPod(t *testing.T) {
	testCases := []struct {
		name     string
		nodeName string
		required bool
		expected *corev1.Affinity
	}{
		{
			name:     "EmptyNodeNameAndPreferredAffinity",
			nodeName: "",
			required: false,
			expected: nil,
		},
		{
			name:     "EmptyNodeNameAndRequiredAffinity",
			nodeName: "",
			required: true,
			expected: nil,
		},
		{
			name:     "NonEmptyNodeNameAndPreferredAffinity",
			nodeName: "node1",
			required: false,
			expected: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 1,
							Preference: corev1.NodeSelectorTerm{
								MatchFields: []corev1.NodeSelectorRequirement{
									{
										Key:      "metadata.name",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"node1"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:     "NonEmptyNodeNameAndPreferredAffinity",
			nodeName: "node1",
			required: true,
			expected: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchFields: []corev1.NodeSelectorRequirement{
									{
										Key:      "metadata.name",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"node1"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, GetNodeAffinityForBackupPod(tc.nodeName, tc.required))
		})
	}
}
