# Running Kimi Agents from Claude Code

*How to dispatch Kimi K2.5 as a subagent from Claude Code*

## Why

Claude Code has a finite context window. Using it for heavy tool work — reading large codebases, running analysis, searching files — displaces relational context. The solution: delegate tool-heavy work to Kimi K2.5 via CLI, keeping Claude’s window for architecture, coordination, and groove.

Kimi K2.5 has a massive quota, excellent code capabilities, and runs via a lightweight CLI.

## Prerequisites

1. **Kimi CLI** installed:  (or )
2. **Config**:  must define a provider and model (see Config section)
3. **API Key**:  env var — get it from [kimi.com](https://kimi.com) or [platform.moonshot.ai](https://platform.moonshot.ai)