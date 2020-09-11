package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jonstacks/jupyterhub-kubernetes-backup/pkg/config"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func fatalIfError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getUserNameFromPVCName(s string) string {
	s = strings.TrimPrefix(s, "claim-")
	for _, encode := range []string{"-20", "-2e"} {
		if strings.Contains(s, encode) {
			s = strings.ReplaceAll(s, encode, ".")
		}
	}
	return s
}

func main() {
	var hasError = false
	var jobsWg sync.WaitGroup

	imageName, ok := os.LookupEnv("BACKUP_IMAGE_NAME")
	if !ok {
		fatalIfError(fmt.Errorf("No BACKUP_IMAGE_NAME variable supplied. Don't know how to launch backup container"))
	}

	clusterConfig, err := rest.InClusterConfig()
	fatalIfError(err)

	clientset, err := kubernetes.NewForConfig(clusterConfig)
	fatalIfError(err)

	namespace := Namespace()

	pvcList, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	fatalIfError(err)

	for _, pvc := range pvcList.Items {
		// Jupyterhub spawns its user data PVCs with "claim-", we are not interested if it doesn't start with this
		if !strings.HasPrefix(pvc.Name, "claim-") {
			continue
		}

		userName := getUserNameFromPVCName(pvc.Name)
		safeUserName := strings.ReplaceAll(userName, ".", "-")
		resourceName := fmt.Sprintf("backup-users-home-%s-%s", pvc.Name, time.Now().Format("200601021504"))

		envVars := []corev1.EnvVar{
			{
				Name:  "LOCAL_PATH",
				Value: "/backup",
			},
			{
				Name:  config.BackupUsername,
				Value: safeUserName,
			},
		}

		copyVars := []string{
			config.Backend,
			config.BackendS3Bucket,
			config.BackendS3Prefix,
			config.BackupUsername,
			config.AwsAccessKeyID,
			config.AwsSecretAccessKey,
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
					},
				},
			},
		}

		log.Printf("Creating job to back up pvc '%s'", pvc.Name)
		resp, err := clientset.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			log.Printf("Err creating backup job for '%s': %s", pvc.Name, err.Error())
			hasError = true
			continue
		}

		log.Printf("Successfully created job '%s' for backing up '%s'\n", resp.Name, pvc.Name)
		jobsWg.Add(1)

		// Now start go-routing to wait and monitor for job to be done...
		go func(jobName string) {
			defer jobsWg.Done()
			defer func() {
				log.Printf("Deleting job '%s'", jobName)
				err := clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, metav1.DeleteOptions{})
				if err != nil {
					log.Printf("Error deleting job '%s': %s", jobName, err)
				}
			}()

			timeout := time.After(10 * time.Minute)

			for {
				select {
				case <-timeout:
					log.Printf("Timeout occured waiting for job to finish")
					return
				default:
					log.Printf("Checking on job '%s' status", jobName)
					resp, err := clientset.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
					if err != nil {
						log.Printf("Error checking job '%s' status: %s", jobName, err)
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
							log.Printf("Job %s failed. Leaving it around for diagnostics", jobName)
							return
						}
					}
				}

				time.Sleep(10 * time.Second)
			}
		}(job.Name)
	}

	if hasError {
		log.Println("Error: Some jobs failed to create")
	}

	jobsWg.Wait()
}
