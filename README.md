# `req` Is a cli tool executing tests in .http files
This tool is inspired by the `http` [extension](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) for Visual Studio Code. It allows you to run tests in `.http` files from the command line.

# Installation

```bash
go install github.com/MordFustang21/req@latest
```

# Usage

```bash
req [flags] <file>
```

# Example file
This is an example of an `.http` file. Multiple tests can be run in a single file. 
More examples and docs can be found at [Here](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)
  
  ```http
  @variable=test-variable
  
  ### Named requests
  GET https://jsonplaceholder.typicode.com/todos/1
  
  // Header section
  Accept: application/json
  
  // Body section
  {
    "userId": 1,
    "id": 1,
    "title": "delectus aut autem",
    "completed": false
  }
  ```

# Goals
- [x] Run tests in `.http` files
- [ ] Log test results
- [ ] Improve UI via bubbletea?
- [ ] Syntax highlighting on output
