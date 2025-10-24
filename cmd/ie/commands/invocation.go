package commands

import (
    "os"
    "github.com/Azure/InnovationEngine/internal/logging"
)

// OriginalInvocationDirectory captures the directory from which the IE command
// was invoked. This allows us to reset the working directory before parsing
// the first document regardless of later changes.
var OriginalInvocationDirectory string

func init() {
    cwd, err := os.Getwd()
    if err != nil {
        logging.GlobalLogger.Warnf("Failed to capture invocation directory: %s", err)
        return
    }
    OriginalInvocationDirectory = cwd
}
