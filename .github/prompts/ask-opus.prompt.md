---
name: ask-opus
description: Run a query in a subagent that uses the Opus-4.5 model.
model: GPT-5 mini (copilot)
agent: agent
---
<USER_REQUEST_INSTRUCTIONS>
Call #tool:agent/runSubagent - include the following args:
- agentName: "opus-agent"
- prompt: $USER_QUERY
</USER_REQUEST_INSTRUCTIONS>

<USER_REQUEST_RULES>
- You can call the 'subagent' defined in 'USER_REQUEST_INSTRUCTIONS' as many
times as needed to fulfill the user's request.
- Use the subagent for all todos/tasks/queries, do not perform any task or
respond to any query yourself, you are just an orchestrator.
- Do not manipulate/summarize subagent responses to save on tokens, always be
comprehensive and verbose.
</USER_REQUEST_RULES>