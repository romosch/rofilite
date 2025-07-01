# Rofilite

Lightweight Rolling File for Go

## Overview
Rofilite is a lightweight rolling file library for Go, designed to provide minimal overhead while managing rotating files efficiently. It is ideal for applications that require simple and effective file rotation without the complexity or overhead of larger libraries.

## Caveat
To reduce system calls, the writer keeps track of the number written bytes to omit checking filesize on every write. The rotation will only occurr after the number of written bytes (not the filesize) has reached the limit. Assuming only one process is writing to the file, this should not be an issue.

## Installation
To install Rofilite, use the following command:

```bash
go get github.com/romosch/rofilite
```

## Usage
Here is a basic example of how to use Rofilite:

```go
package main

import (
    "github.com/romosch/rofilite"
)

func main() {
    logger, err := rofilite.New("app.log",
        rofilite.WithMaxBytes(10 * 1024 * 1024), // 10 MB
        rofilite.WithMaxBackups(5),
    )
    if err != nil {
        panic(err)
    }

    logger.Write([]byte("Hello!"))
    logger.Close()
}
```

## Configuration

Rofilite provides several options to customize the behavior of the rolling file:

- `WithMaxBytes(maxBytes int64)`: Sets the maximum size in bytes before the log file is rotated.
- `WithMaxBackups(maxBackups int)`: Specifies the maximum number of backup files to retain.
- `WithMaxAge(age time.Duration)`: Defines the maximum age of backup files before they are deleted.
- `WithMode(mode os.FileMode)`: Sets the file mode for the log file on creation.
- `WithErrorHandler(handler func(error))`: Allows setting a custom handler for errors occurring during the cleanup of backup files.

## Contributing
Contributions are welcome! Feel free to open issues or submit pull requests to improve the library.

## License
This project is licensed under the MIT License. See the LICENSE file for details.