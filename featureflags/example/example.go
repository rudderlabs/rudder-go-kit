package main

import (
	"fmt"
	"log"

	"github.com/rudderlabs/rudder-go-kit/featureflags"
)

func main() {
	featureflags.SetDefaultTraits(map[string]string{"tier": "ENTERPRISE_V1"})
	isenabled, err := featureflags.IsFeatureEnabled("entTest", "enterpriseonlyfeature")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(isenabled)
	
}
