# Plugin System

The plugin system allows you to extend aidev with custom commands, AI providers, and integrations. Plugins are standard Node.js packages that hook into aidev's lifecycle.

---

## Overview

Plugins can:

- **Add new slash commands** -- Create custom `/my-command` slash commands
- **Register AI providers** -- Add support for new AI models or API endpoints
- **Add middleware** -- Intercept and modify requests/responses
- **Register hooks** -- Run code at specific points in the aidev lifecycle
- **Extend the UI** -- Add custom output renderers or theme components

---

## Managing Plugins

### Install a Plugin

```bash
# From npm
aidev plugins install aidev-plugin-docker

# From a local path
aidev plugins install ./my-plugin

# From a Git repository
aidev plugins install github:user/aidev-plugin-terraform
```

### List Installed Plugins

```bash
aidev plugins list
```

Output:
```
Installed Plugins:
  aidev-plugin-docker    v1.2.0  - Docker and Docker Compose integration
  aidev-plugin-terraform v0.5.0  - Terraform plan/apply assistance
  my-plugin              v1.0.0  - Custom team workflows
```

### Enable / Disable Plugins

```bash
# Disable a plugin without uninstalling
aidev plugins disable aidev-plugin-docker

# Re-enable a plugin
aidev plugins enable aidev-plugin-docker
```

### Uninstall a Plugin

```bash
aidev plugins uninstall aidev-plugin-docker
```

---

## Creating a Plugin

### Scaffold a New Plugin

```bash
aidev plugins create my-plugin
```

This creates the following structure:

```
my-plugin/
+-- package.json
+-- src/
|   +-- index.ts        # Plugin entry point
|   +-- commands/       # Custom commands
|   +-- hooks/          # Lifecycle hooks
+-- README.md
+-- tsconfig.json
```

### Plugin Entry Point

The main entry point exports a plugin definition object:

```typescript
// src/index.ts
import { AidevPlugin } from 'aidev-cli';

const plugin: AidevPlugin = {
  name: 'my-plugin',
  version: '1.0.0',
  description: 'My custom plugin',

  // Called when the plugin is loaded
  async initialize(context) {
    context.logger.info('my-plugin initialized');
  },

  // Register slash commands
  commands: [
    {
      name: 'deploy',
      description: 'Deploy the current project',
      aliases: ['d'],
      handler: async (args, context) => {
        const target = args[0] || 'staging';
        context.logger.info(`Deploying to ${target}...`);
        // Implementation here
      },
    },
  ],

  // Register hooks
  hooks: {
    // Before sending a message to the AI provider
    'message:before': async (message, context) => {
      // Modify or enrich the message
      return message;
    },

    // After receiving a response from the AI provider
    'message:after': async (response, context) => {
      // Process the response
      return response;
    },

    // Before executing a tool/function call
    'tool:before': async (tool, args, context) => {
      context.logger.debug(`Executing tool: ${tool.name}`);
      return { tool, args };
    },

    // After the REPL starts
    'app:ready': async (context) => {
      context.logger.info('aidev is ready');
    },

    // Before exiting
    'app:exit': async (context) => {
      // Cleanup
    },
  },

  // Register middleware (runs in order)
  middleware: [
    {
      name: 'logging',
      order: 10,
      handler: async (request, next) => {
        const start = Date.now();
        const response = await next(request);
        const duration = Date.now() - start;
        console.log(`Request took ${duration}ms`);
        return response;
      },
    },
  ],

  // Register a custom AI provider
  providers: [
    {
      id: 'my-provider',
      name: 'My Custom Provider',
      async chat(messages, options) {
        // Implement chat completion
        return {
          content: 'Response from my provider',
          usage: { inputTokens: 100, outputTokens: 50 },
        };
      },
      async *stream(messages, options) {
        // Implement streaming
        yield { type: 'text', content: 'Streaming ' };
        yield { type: 'text', content: 'response' };
        yield { type: 'done' };
      },
    },
  ],
};

export default plugin;
```

---

## Plugin Context

The `context` object passed to plugin handlers provides access to aidev internals:

| Property | Type | Description |
|---|---|---|
| `context.logger` | `Logger` | Logging utility (`info`, `warn`, `error`, `debug`) |
| `context.config` | `Config` | Access to the merged configuration |
| `context.session` | `Session` | Current conversation session |
| `context.history` | `History` | Conversation history manager |
| `context.git` | `GitTools` | Git operations |
| `context.ui` | `UI` | Terminal UI utilities |
| `context.store` | `Store` | Key-value store for plugin state |

### Plugin Store

Plugins get a persistent key-value store:

```typescript
// Save data
await context.store.set('lastDeployment', {
  target: 'production',
  timestamp: new Date().toISOString(),
});

// Read data
const last = await context.store.get('lastDeployment');

// Delete data
await context.store.delete('lastDeployment');
```

---

## Available Hooks

| Hook | Trigger | Arguments |
|---|---|---|
| `app:ready` | After REPL initialization | `context` |
| `app:exit` | Before process exit | `context` |
| `message:before` | Before sending to AI provider | `message, context` |
| `message:after` | After receiving AI response | `response, context` |
| `tool:before` | Before tool execution | `tool, args, context` |
| `tool:after` | After tool execution | `tool, result, context` |
| `config:change` | When config is modified | `key, value, context` |
| `session:start` | New session started | `session, context` |
| `session:end` | Session ended | `session, context` |
| `error` | On unhandled error | `error, context` |

---

## Publishing a Plugin

### Package Requirements

Your `package.json` must include:

```json
{
  "name": "aidev-plugin-my-feature",
  "version": "1.0.0",
  "description": "Description of what this plugin does",
  "main": "dist/index.js",
  "types": "dist/index.d.ts",
  "keywords": ["aidev", "aidev-plugin", "ai", "cli"],
  "peerDependencies": {
    "aidev-cli": ">=1.0.0"
  },
  "aidev": {
    "minVersion": "1.0.0"
  }
}
```

### Naming Convention

- Plugin packages should be prefixed with `aidev-plugin-`
- Example: `aidev-plugin-docker`, `aidev-plugin-terraform`

### Publishing to npm

```bash
npm login
npm publish
```

After publishing, users can install with:

```bash
aidev plugins install aidev-plugin-my-feature
```

---

## Example Plugins

### Custom Deployment Command

```typescript
const plugin: AidevPlugin = {
  name: 'deploy',
  version: '1.0.0',
  description: 'Deploy to cloud platforms',

  commands: [
    {
      name: 'deploy',
      description: 'Deploy current project',
      handler: async (args, ctx) => {
        const platform = args[0] || 'vercel';
        ctx.ui.spinner(`Deploying to ${platform}...`);

        const { execSync } = await import('child_process');
        try {
          execSync(`${platform} deploy`, { stdio: 'inherit' });
          ctx.ui.success('Deployment complete!');
        } catch (err) {
          ctx.ui.error('Deployment failed');
        }
      },
    },
  ],
};

export default plugin;
```

### Custom Provider with Auth

```typescript
const plugin: AidevPlugin = {
  name: 'custom-llm',
  version: '1.0.0',
  description: 'Connect to a custom LLM endpoint',

  providers: [
    {
      id: 'custom-llm',
      name: 'Custom LLM',
      requiresApiKey: true,

      async chat(messages, options) {
        const response = await fetch(options.baseUrl + '/chat', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${options.apiKey}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ messages }),
        });
        return response.json();
      },
    },
  ],
};

export default plugin;
```

---

## Troubleshooting

**Plugin not loading:**
Run `aidev plugins list` to check if the plugin is installed and enabled. Check `aidev doctor` for any dependency conflicts.

**Version incompatibility:**
Ensure the plugin's `peerDependencies` match your aidev version. Update with `aidev plugins update`.

**Plugin conflicts:**
If two plugins register the same command name, the one installed later takes precedence. Use `aidev plugins disable` to resolve conflicts.
