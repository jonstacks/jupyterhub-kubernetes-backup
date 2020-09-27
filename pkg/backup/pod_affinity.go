package backup

import (
	corev1 "k8s.io/api/core/v1"
)

// GetNodeAffinityForBackupPod returns an Affinity for the backup pod given the
// nodeName that the user's jupyter pod is running on. If required is true, a
// NodeAffinity with a RequiredDuringSchedulingIgnoredDuringExecution will be
// returned, otherwise a NodeAffinity with
// PreferredDuringSchedulingIgnoredDuringExecution will be returned.
func GetNodeAffinityForBackupPod(nodeName string, required bool) *corev1.Affinity {
	if nodeName == "" {
		return nil
	}

	nodeSelectorTerm := corev1.NodeSelectorTerm{
		// Add a match field which matches the node's metadata.name field to the
		// nodename the pod is running on.
		MatchFields: []corev1.NodeSelectorRequirement{
			{
				Key:      "metadata.name",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{nodeName},
			},
		},
	}

	if required {
		return &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						nodeSelectorTerm,
					},
				},
			},
		}
	}

	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight:     1,
					Preference: nodeSelectorTerm,
				},
			},
		},
	}
}
