# agfapi
A command‑line and Go library tool for interacting with the AGFA FHIR API.

## Table of Contents
1. [Overview](#overview)
2. [Features](#features) 
3. [Installation](#installation)
4. [Usage](#usage)
   - CLI tool
   - Library usage
5. [Configuration & Environment](#configuration‑environment)
5. [Contributing](#contributing)
7. [License](#license)

## Overview
agfapi is a Go‑based tool that provides both a CLI and a library for interacting with an AGFA FHIR service. The aim is to simplify making API calls, handle authentication, and integrate with your Go applications or pipelines.

## Features
- Command‑line interface for quick use.
- Library package that you can import into your Go code.

## Installation

### TODO: include binary installs

### Go (requires 1.25.0+)
```bash
go install github.com/s‑hammon/agfapi/cmd/agfapi@latest
```

### Build from source
```bash
git clone https://github.com/s‑hammon/agfapi.git
cd agfapi
make build
```

## Usage

WIP

## Configuration & Environment
The base URL for your AGFA service should be assigned to an `AGFA_URL` environment variable:

```bash
export AGFA_URL=https://your.agfa-url.com/fhir/r4
```

## Contributing
Contributions are welcome! If you would like to help, please fork the repository and create a branch for your feature/bugfix.

## License
This project is currently unlicensed (**Note:** please add `LICENSE` file if needed).
If you adopt a license, specify it here (e.g., MIT, Apache 2.0).

