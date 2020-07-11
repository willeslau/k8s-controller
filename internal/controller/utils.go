package controller

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func createSelectorsFromLabels(ls map[string]string) (labels.Selector, error) {
	selector := labels.NewSelector()
	for key, val := range ls {
		r, err := labels.NewRequirement(key, selection.Equals, []string{val})
		if err != nil {
			return nil, err
		}
		selector = selector.Add(*r)
	}
	return selector, nil
}
