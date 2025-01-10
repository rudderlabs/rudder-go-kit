
# Feature Flags Package

A Go package for managing feature flags using Flagsmith as the provider.

## Prerequisites

Before using this package, ensure you have set the following environment variable:

```bash
export FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY=your-api-key
```

## Usage Example

Here's a basic example of how to use the feature flags package:

```go
package main

import (
    "fmt"
    "log"

    "github.com/rudderlabs/rudder-go-kit/featureflags"
)

func main() {
    // Set default traits (optional)
    featureflags.SetDefaultTraits(map[string]string{
        "exampleTrait": "exampleTraitValue",
    })

    // Check if a feature is enabled
    isEnabled, err := featureflags.IsFeatureEnabled("testWs", "testFeature")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(isEnabled)
}
```

For more detailed examples and implementation details, refer to the example directory.