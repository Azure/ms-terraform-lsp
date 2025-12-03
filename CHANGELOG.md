## v0.8.0

Enhancements:
- Improve validation logic: ignore `readonly` + `required` properties if missing or defined.
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/c41a40c0d2f9fa78b7ea0901b6634a13dc8e8b33
- Update aztfmigrate to v2.8.0 to support azurerm v4.54.0 resources.


## v0.7.0

Enhancements:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/6b9ca69c973d29e6cf745cb5f25b13ee033de985

## v0.6.0
Features:
- Support auto-completion and hover documentation for Azure Verified Modules.
- Add auto-completion and hover documentation for data sources.

Enhancements:
- Update bicep types to https://github.com/ms-henglu/bicep-types-az/commit/a3cf29cb316d792abe0a607f97469a577382ee77

Bug Fixes:
- Correct hover documentation snippet formatting.
- Fix data source markdown generation script.

## v0.5.0
Features
- Support language features for `azapi` provider resources and data sources.
- Support `refactor/rewrite` code action which can trigger the command to convert resources between `azapi` and `azurerm` providers.
- Support `aztfmigrate` command which can convert resources between `azapi` and `azurerm` providers.
- Support `workspace/executeCommand` protocol which can convert ARMTemplate and resource JSON content to azapi configuration.
- Support generating required/missing permissions for `azapi` provider resources and data sources.

## v0.4.0
Enhancements:
- Improve the error messages for authentication issues.

Bug Fixes:
- Fix the bug that incorrect hover documentation and completion suggestions were shown for `azurerm` data sources.

## v0.3.0
Features:
- Rename packages to `ms-terraform-lsp`

Bug Fixes:
- Fix the msgraph memeber resource template to use the correct url value.

## v0.2.0

Features
- Support auto-completion and hover documentation for `msgraph` provider resources and data sources.

## v0.1.0

Features:
- Support auto-completion and hover documentation for `azurerm` provider resources and data sources.
- Support generating required/missing permissions for `azurerm` provider resources and data sources.

