# Data Toolkit V1

Data Toolkit is a deterministic, local, extensible data-processing product.
This repository currently contains the V1 product foundation: canonical data
contracts, explicit extension registries, orchestration, a shared application
API, strict configuration, and interface scaffolds.

The foundation intentionally ships no end-user data operations or file-format
implementations. JSON, CSV, Excel, TXT, and Google packages are empty extension
scaffolds until separate OpenSpec changes define their product behavior.

## Verify

```powershell
.\scripts\dev.ps1
```

The script formats the V1 module, runs tests and vet, builds the executable, and
strictly validates the active V1 OpenSpec change.

## Discover the compiled surface

```powershell
go run ./cmd/data-toolkit capabilities
go run ./cmd/data-toolkit config validate ./configs/default.yaml
```

The partial predecessor is retained under `legacy/mvp` as an isolated nested Go
module. It is not part of V1 runtime or root quality gates.
