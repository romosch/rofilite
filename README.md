# rollingfile

Rolling File (Light) for Go

## Overview
rollingfile is a lightweight rolling file library for Go, designed to incur minimal overhead while managing rotating files. It is ideal for applications that require simple and effective file rotation without the complexity or overhead of larger libraries.

## Features
### Minimal System Calls
Unlike other libraries, the writer keeps track of the number written bytes to omit checking filesize on every write. The rotation will only occurr after the number of written bytes (not the current filesize) has reached the limit. This parts from the assumption only one process/goroutine will be writing to the file.

### Non-blocking Cleanup 
The cleanup of backup files (according to values defined in `WithMaxAge` or `WithMaxBackups`) is performed in an additional goroutine to reduce the time a call to `Write` waits for a file-rotation to complete. Errors occurring during cleanup can be handled by a custom function passed via the `WithErrorHandler` option

## Installation
To install rollingfile, use the following command:

```bash
go get github.com/romosch/rollingfile
```

## Usage
Here is a basic example of how to use rollingfile:

```go
package main

import (
    "github.com/romosch/rollingfile"
)

func main() {
    logger, err := rollingfile.New("app.log",
        rollingfile.WithMaxBytes(10 * 1024 * 1024), // 10 MB
        rollingfile.WithMaxBackups(5),
    )
    if err != nil {
        panic(err)
    }

    logger.Write([]byte("Hello!"))
    logger.Close()
}
```

## Configuration

rollingfile provides several options to customize the behavior of the rolling file:

- `WithMaxBytes(maxBytes int64)`: Sets the maximum size in bytes before the log file is rotated.
- `WithMaxBackups(maxBackups int)`: Specifies the maximum number of backup files to retain.
- `WithMaxAge(age time.Duration)`: Defines the maximum age of backup files before they are deleted.
- `WithMode(mode os.FileMode)`: Sets the file mode for the log file on creation.
- `WithErrorHandler(handler func(error))`: Allows setting a custom handler for errors occurring during the cleanup of backup files.

## Contributing
Contributions are welcome! Feel free to open issues or submit pull requests to improve the library.

## License
This project is licensed under the MIT License. See the LICENSE file for details.