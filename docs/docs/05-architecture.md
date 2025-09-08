---
title: Architecture
---

## Overview

The Obot platform is designed to enable users to consume MCP servers in an enterprise setting. It consists of three main components that work together to provide a complete solution: the MCP Gateway for tool integration, the Chat Interface for user interaction, and the Admin Interface for platform management.

![ALT Text](/img/high-level-arch.png)

## Concepts

- **MCP Servers:** MCP servers are the bits of code that implement the MCP spec like tools, prompts, and resources. It serves them up to be consumed by the clients to connect with LLMs.

- **MCP Clients:** MCP clients are the tools that interact with LLMs and consume MCP servers. These clients can be agents, desktop tools like Cursor, Claude, VS Code, or any other applications that interface with LLMs. They are responsible for handling the interactions with MCP servers, LLMs, and the end user.

- **MCP Catalog:** The catalog is an index of MCP servers that contains metadata about them, including how to run them and where to get them. It serves as a centralized repository for discovering and managing available MCP servers.

- **MCP Gateway:** The gateway is a proxy that sits between the client and the server, providing important capabilities for enterprise deployment. It performs auditing, inspects payloads, and applies policy and access controls to MCP servers. This is a critical piece of the tool and capability infrastructure.

- **LLM Gateway:** The LLM gateway (or LLM proxy) is the connectivity layer that sits between a chat client and the LLM. It also provides proxy capabilities to the interaction, enabling additional control and monitoring of LLM communications.