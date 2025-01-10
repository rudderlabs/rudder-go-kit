package main

import (
	"fmt"
	"log"

	"github.com/rudderlabs/rudder-go-kit/featureflags"
)

func main() {
	featureflags.SetDefaultTraits(map[string]string{"exampleTrait": "exampleTraitValue"})
	isenabled, err := featureflags.IsFeatureEnabled("testWs", "testFeature")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(isenabled)
}
