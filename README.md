# Microsoft Terraform Providers Language Server

Experimental version of Microsoft Terraform Providers language server.

## What is LSP

Read more about the Language Server Protocol at https://microsoft.github.io/language-server-protocol/

## Introduction

This project only supports language features for Microsoft Terraform providers,
not targeting support all language features for `HCL` or `Terraform`. To get the best user experience, 
it's recommended to use it with language server for `Terraform`.

## Installation

1. Clone this project to local
2. Run `go install` under the project folder.

## Usage

The most reasonable way you will interact with the language server
is through a client represented by an IDE, or a plugin of an IDE.

VSCode extension: [vscode-azureterraform](https://github.com/Azure/vscode-azureterraform)

## Credits

We wish to thank HashiCorp for the use of some MPLv2-licensed code from their open source project [terraform-ls](https://github.com/hashicorp/terraform-ls).

