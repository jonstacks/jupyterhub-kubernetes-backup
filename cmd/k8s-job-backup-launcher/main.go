package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/backup"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/config"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/core"
	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/k8scontrib"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	logrus.SetFormatter(config.GetLogFormatter())
	logrus.SetLevel(config.GetLogLevel())
}

func cleanupSuccessfulBackupPods(clientset *kubernetes.Clientset, namespace string) {
	logrus.Info("Cleaning up pods that have exited successfully")

	resp, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("Unable to list pods to clean up: %s", err)
	}

	for _, pod := range resp.Items {
		if !strings.HasPrefix(pod.Name, "backup-users-home") {
			continue
		}
		if pod.Status.Phase != corev1.PodSucceeded {
			continue
		}
		logrus.Debugf("Deleting pod '%s'", pod.Name)
		err = clientset.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			logrus.Errorf("Unable to delete pod '%s'", pod.Name)
			continue
		}
		logrus.Infof("Successfully deleted pod '%s'", pod.Name)
	}
}

func main() {
	var hasError = false
	var jobsWg sync.WaitGroup

	imageName, ok := os.LookupEnv("BACKUP_IMAGE_NAME")
	if !ok {
		core.FatalIfError(fmt.Errorf("No BACKUP_IMAGE_NAME variable supplied. Don't know how to launch backup container"))
	}

	clusterConfig, err := rest.InClusterConfig()
	core.FatalIfError(err)

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	core.FatalIfError(err)

	namespace := k8scontrib.Namespace()

	pvcList, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	core.FatalIfError(err)

	for _, pvc := range pvcList.Items {
		// Jupyterhub spawns its user data PVCs with "claim-", we are not interested if it doesn't start with this
		if !strings.HasPrefix(pvc.Name, "claim-") {
			logrus.Debugf("Skipping pvc '%s'", pvc.Name)
			continue
		}

		// If a pod is running, this is the name we'd expect to find
		podName := strings.Replace(pvc.Name, "claim", "jupyter", 1)
		var affinity *corev1.Affinity
		var tolerations []corev1.Toleration

		logrus.Infof("Searching for pod with name '%s'", podName)
		pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			logrus.Debugf("Error searching for pod '%s': %s", podName, err)
		} else if pod != nil {
			logrus.Infof("Found currently running pod '%s' for pvc '%s'. Adding node affinity.", pod.Name, pvc.Name)
			if pod.Spec.NodeName != "" {
				affinity = backup.GetNodeAffinityForBackupPod(pod.Spec.NodeName, config.GetBackupPodNodeAffinityRequired())
			}
			// Copy the tolerations from the user's pod to the backup pod.
			tolerations = pod.Spec.Tolerations
		} else {
			logrus.Debugf("No pod found with name '%s'", podName)
		}

		userName := backup.GetUserNameFromPVCName(pvc.Name)
		safeUserName := strings.ReplaceAll(userName, ".", "-")
		resourceName := fmt.Sprintf("backup-users-home-%s-%s", pvc.Name, time.Now().Format("200601021504"))

		envVars := []corev1.EnvVar{
			{
				Name:  "LOCAL_PATH",
				Value: "/backup",
			},
			{
				Name:  config.BackupUsername,
				Value: userName,
			},
		}

		copyVars := []string{
			config.Backend,
			config.BackendS3Bucket,
			config.BackendS3Prefix,
			config.BackupUsername,
			config.BackupPodNodeAffinity,
			config.AwsAccessKeyID,
			config.AwsSecretAccessKey,
			config.AwsDefaultRegion,
			config.LogLevel,
			config.LogFormat,
		}

		for _, name := range copyVars {
			if val, ok := os.LookupEnv(name); ok {
				envVars = append(envVars, corev1.EnvVar{
					Name:  name,
					Value: val,
				})
			}
		}

		// Now launch a job to back up the user's directory
		job := &batchv1.Job{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   resourceName,
				Labels: make(map[string]string),
			},
			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:   resourceName,
						Labels: make(map[string]string),
					},
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{},
						Containers: []corev1.Container{
							{
								Name:            safeUserName,
								Image:           imageName,
								Command:         []string{"/usr/local/bin/jupyterhub-kubernetes-backup"},
								ImagePullPolicy: corev1.PullAlways,
								Env:             envVars,
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "user-backup",
										MountPath: "/backup",
										ReadOnly:  true,
									},
								},
							},
						},
						RestartPolicy:    corev1.RestartPolicyNever,
						ImagePullSecrets: []corev1.LocalObjectReference{},
						Volumes: []corev1.Volume{
							{
								Name: "user-backup",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: pvc.Name,
										ReadOnly:  true,
									},
								},
							},
						},
						Affinity:    affinity,
						Tolerations: tolerations,
					},
				},
			},
		}

		logrus.Infof("Creating job to back up pvc '%s'", pvc.Name)
		resp, err := clientset.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			logrus.Errorf("Err creating backup job for '%s': %s", pvc.Name, err.Error())
			hasError = true
			continue
		}

		logrus.Infof("Successfully created job '%s' for backing up '%s'\n", resp.Name, pvc.Name)
		jobsWg.Add(1)

		// Now start go-routing to wait and monitor for job to be done...
		go func(jobName string) {
			defer jobsWg.Done()
			defer func() {
				logrus.Infof("Deleting job '%s'", jobName)
				err := clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
				if err != nil {
					logrus.Errorf("Error deleting job '%s': %s", jobName, err)
				}
			}()

			timeout := time.After(10 * time.Minute)

			for {
				select {
				case <-timeout:
					logrus.Errorf("Timeout occured waiting for job to finish")
					return
				default:
					logrus.Infof("Checking on job '%s' status", jobName)
					resp, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
					if err != nil {
						logrus.Errorf("Error checking job '%s' status: %s", jobName, err)
						time.Sleep(30 * time.Second)
						continue
					}

					// Job was successful
					if resp.Status.CompletionTime != nil {
						time.Sleep(30 * time.Second)
						return
					}

					for _, cond := range resp.Status.Conditions {
						if cond.Type == batchv1.JobFailed {
							logrus.Infof("Job %s failed. Leaving it around for diagnostics", jobName)
							return
						}
					}
				}

				time.Sleep(10 * time.Second)
			}
		}(job.Name)
	}

	if hasError {
		logrus.Errorf("Error: Some jobs failed to create")
	}

	jobsWg.Wait()

	cleanupSuccessfulBackupPods(clientset, namespace)
}
