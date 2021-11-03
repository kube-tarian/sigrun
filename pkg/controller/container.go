package controller

import (
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/api/admission/v1beta1"
)

// getContainersFromPodOrDeployment returns the containers from a kubernetes object
func GetContainersFromResource(req *v1beta1.AdmissionReview) ([]corev1.Container, error) {

	var containers []corev1.Container
	switch req.Request.Kind.Kind {
	case "Pod":
		var pod corev1.Pod
		err := json.Unmarshal(req.Request.Object.Raw, &pod)
		if err != nil {
			return nil, NewError("failed to unmarshal pod", err)
		}
		containers = getContainersFromPod(pod)

	case "Deployment":
		var deployment appsv1.Deployment
		err := json.Unmarshal(req.Request.Object.Raw, &deployment)
		if err != nil {
			return nil, NewError("failed to unmarshal deployment", err)
		}
		containers = getContainersFromDeployment(deployment)
	case "StatefulSet":
		var statefulSet appsv1.StatefulSet
		err := json.Unmarshal(req.Request.Object.Raw, &statefulSet)
		if err != nil {
			return nil, NewError("failed to unmarshal statefulset", err)
		}

		containers = getContainersFromStatefulSet(statefulSet)
	case "DaemonSet":
		var daemonSet appsv1.DaemonSet
		err := json.Unmarshal(req.Request.Object.Raw, &daemonSet)
		if err != nil {
			return nil, NewError("failed to unmarshal daemonset", err)
		}

		containers = getContainersFromDaemonSet(daemonSet)
	case "ReplicaSet":
		var replicaSet appsv1.ReplicaSet
		err := json.Unmarshal(req.Request.Object.Raw, &replicaSet)
		if err != nil {
			return nil, NewError("failed to unmarshal replicaset", err)
		}

		containers = getContainersFromReplicaSet(replicaSet)
	case "ReplicationController":
		var replicationController corev1.ReplicationController
		err := json.Unmarshal(req.Request.Object.Raw, &replicationController)
		if err != nil {
			return nil, NewError("failed to unmarshal replicationcontroller", err)
		}

		containers = getContainersFromReplicationController(replicationController)
	case "Job":
		var job batchv1.Job
		err := json.Unmarshal(req.Request.Object.Raw, &job)
		if err != nil {
			return nil, NewError("failed to unmarshal job", err)
		}

		containers = getContainersFromJob(job)
	case "CronJob":
		var cronJob batchv1.CronJob
		err := json.Unmarshal(req.Request.Object.Raw, &cronJob)
		if err != nil {
			return nil, NewError("failed to unmarshal cronjob", err)
		}

		containers = getContainersFromCronJob(cronJob)

	case "Container":
		var container corev1.Container
		err := json.Unmarshal(req.Request.Object.Raw, &container)
		if err != nil {
			return nil, NewError("failed to unmarshal container", err)
		}

		containers = append(containers, container)
	default:
		return nil, NewError("unsupported kind", nil)
	}

	return containers, nil
}

func getContainersFromCronJob(cronJob batchv1.CronJob) []corev1.Container {
	containers := append(cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers, cronJob.Spec.JobTemplate.Spec.Template.Spec.InitContainers...)
	return containers
}

func getContainersFromJob(job batchv1.Job) []corev1.Container {
	containers := append(job.Spec.Template.Spec.Containers, job.Spec.Template.Spec.InitContainers...)
	return containers
}

func getContainersFromReplicaSet(replicaSet appsv1.ReplicaSet) []corev1.Container {
	containers := append(replicaSet.Spec.Template.Spec.Containers, replicaSet.Spec.Template.Spec.InitContainers...)
	return containers
}

func getContainersFromReplicationController(replicationController corev1.ReplicationController) []corev1.Container {
	containers := append(replicationController.Spec.Template.Spec.Containers, replicationController.Spec.Template.Spec.InitContainers...)
	return containers

}

func getContainersFromDaemonSet(daemonSet appsv1.DaemonSet) []corev1.Container {
	containers := append(daemonSet.Spec.Template.Spec.Containers, daemonSet.Spec.Template.Spec.InitContainers...)
	return containers
}

func getContainersFromStatefulSet(statefulSet appsv1.StatefulSet) []corev1.Container {
	containers := append(statefulSet.Spec.Template.Spec.Containers, statefulSet.Spec.Template.Spec.InitContainers...)
	return containers
}

func getContainersFromDeployment(deployment appsv1.Deployment) []corev1.Container {
	containers := append(deployment.Spec.Template.Spec.Containers, deployment.Spec.Template.Spec.InitContainers...)
	return containers
}

func getContainersFromPod(pod corev1.Pod) []corev1.Container {
	containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)
	return containers
}
