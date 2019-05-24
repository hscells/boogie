package main

import (
	"bytes"
	"fmt"
	"github.com/hscells/groove/learning"
	"os"
	"sync"
)

func main() {
	fmt.Println("loading features")
	lf, err := learning.LoadFeatures(os.Stdin)
	if err != nil {
		panic(err)
	}
	fmt.Println("loading model")
	s := learning.NewNearestNeighbourCandidateSelector(learning.NearestNeighbourLoadModel(os.Args[1]), learning.NearestNeighbourDepth(10))
	m := make(map[string]float64)
	q := make(map[string]learning.LearntFeature)
	fmt.Println("all set!")
	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for _, f := range lf {
		wg.Add(1)
		go func(feature learning.LearntFeature) {
			defer wg.Done()
			score := s.Predict(feature)
			mu.Lock()
			defer mu.Unlock()
			if _, ok := m[feature.Topic]; !ok {
				m[feature.Topic] = 0.0
			}
			if score > m[feature.Topic] {
				fmt.Printf("[%s] %f *\n", feature.Topic, score)
				m[feature.Topic] = score
				q[feature.Topic] = feature
			} else {
				fmt.Printf("[%s] %f\n", feature.Topic, score)
			}
		}(f)
	}
	wg.Wait()
	buff := new(bytes.Buffer)
	for k, v := range q {
		buff.WriteString(fmt.Sprintf("%f %s %s\n", m[k], k, v.Comment))
	}
	f, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)
	if err != nil {
		panic(err)
	}
	_, err = buff.WriteTo(f)
	if err != nil {
		panic(err)
	}
}
