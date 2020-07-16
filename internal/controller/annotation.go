package controller

import (
	v1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	appsv1 "k8s.io/api/apps/v1"
)

type annotation string

const (
	SPEC_HASH annotation = "willesxm.k8s.io/spec-hash"
	REVISION_VERSION annotation = "willesxm.k8s.io/revision"
)

var DEFAULT_ANNOTATIONS = []annotation{SPEC_HASH, REVISION_VERSION}

type annotations struct {
	anns map[string]string
}

// Set will set annotation to the annotations
func (a * annotations) Set(ann annotation, val string) {
	a.anns[string(ann)] = val
}

// Del will delete annotation to the annotations
func (a * annotations) Del(ann annotation) {
	delete(a.anns, string(ann))
}

// Get will get annotation from the annotations
func (a * annotations) Get(ann annotation) string {
	if val, ok := a.anns[string(ann)]; ok {
		return val
	}
	return ""
}

//// CopyFieldsFrom will delete annotation to the annotations
//func (a * annotations) CopyFieldsFrom(p *annotations, fields []annotation) {
//	if fields == nil { fields = DEFAULT_ANNOTATIONS }
//	for _, field := range fields {
//		if len(p.Get(field)) == 0 { continue }
//		a.Set(field, p.Get(field))
//	}
//}

// CopyFrom will delete annotation to the annotations
func (a * annotations) CopyFrom(p *annotations) bool {
	updated := false
	for _, field := range DEFAULT_ANNOTATIONS {
		if len(p.Get(field)) == 0 || a.Get(field) == p.Get(field) { continue }
		a.Set(field, p.Get(field))
		updated = true
	}
	return updated
}

// dump will dump the annotations
func (a * annotations) dump() *map[string]string {
	return &a.anns
}

// AssignToWorker assigns the annotation to deployment
func (a * annotations) SetDeploymentAnnotations(deployment *appsv1.Deployment) {
	deployment.Annotations = *a.dump()
}

// AssignToWorker assigns the annotation to worker
func (a * annotations) AssignToWorker(worker *v1.Worker) {
	worker.Annotations = *a.dump()
}

// newAnnotation will create a new annotations struct
func newAnnotation() *annotations {
	return &annotations{make(map[string]string)}
}

// newAnnotation will create a new annotations struct from map
func newAnnotationFromMap(m map[string]string) *annotations {
	a := make(map[string]string)
	for k, v := range m {
		a[k] = v
	}
	return &annotations{a}
}
