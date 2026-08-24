# Microsoft Terraform Providers Language Server

Experimental version of Microsoft Terraform Providers language server.

## What is LSP

Read more about the Language Server Protocol at <https://microsoft.github.io/language-server-protocol/>

## Introduction

This project only supports language features for Microsoft Terraform providers,
not targeting support all language features for `HCL` or `Terraform`. To get the best user experience,
it's recommended to use it with language server for `Terraform`.

## Installation

Download a released binary for your platform from [Releases](https://github.com/Azure/ms-terraform-lsp/releases), or install with Go:

```
go install github.com/Azure/ms-terraform-lsp@latest
```

To build from source, clone this repository and run `go install` in the project folder.

## Usage

The most reasonable way you will interact with the language server
is through a client represented by an IDE, or a plugin of an IDE.

VSCode extension: [vscode-azureterraform](https://github.com/Azure/vscode-azureterraform)

### Other editors

The language server communicates over stdio and is started with the `serve` subcommand, so any LSP client can drive it:

```
ms-terraform-lsp serve
```

It registers for the `terraform` language and takes no LSP `settings`.

Example configuration for Neovim 0.11+, using the config that ships with
[nvim-lspconfig](https://github.com/neovim/nvim-lspconfig/blob/master/lsp/ms_terraform_lsp.lua):

```lua
vim.lsp.enable('ms_terraform_lsp')
```

Neovim users can install the binary with [mason.nvim](https://github.com/mason-org/mason.nvim):

```
:MasonInstall ms-terraform-lsp
```

## Credits

We wish to thank HashiCorp for the use of some MPLv2-licensed code from their open source project [terraform-ls](https://github.com/hashicorp/terraform-ls).
