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

