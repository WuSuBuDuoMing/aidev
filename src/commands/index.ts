/**
 * aidev CLI — Slash Commands
 *
 * Handles all /command inputs from the user during an interactive session.
 * Each command is registered with a name, description, and handler.
 *
 * @module commands
 */

import * as readline from 'readline';
import chalk from 'chalk';
import { AppConfig, AIProviderEnum, Conversation, Message, PartialConfig, Skill } from '../types';
import { ToolRegistry } from '../tools';
import { ProjectContext } from '../types';
import { renderMarkdown, setTheme, getTheme, statusBar, horizontalRule, getColors } from '../ui/terminal';
import {
  codingAssistantPrompt,
  codeReviewPrompt,
  refactorPrompt,
  explainPrompt,
  testGenerationPrompt,
  commitMessagePrompt,
  prDescriptionPrompt,
} from '../ai/prompts';
import { createProvider, defaultModel, defaultBaseUrl } from '../ai/provider';
import { saveGlobalConfig, validateConfig } from '../config';
import { loadSession, listSessions } from '../storage/session';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/** Context passed to every command handler. */
export interface CommandContext {
  config: AppConfig;
  conversation: Conversation;
  projectContext: ProjectContext;
  toolRegistry: ToolRegistry;
  skills: Skill[];
  /** Update the active config (partial merge). */
  updateConfig(patch: PartialConfig): void;
  /** Start a fresh conversation. */
  newConversation(): void;
  /** Reload project context. */
  refreshContext(): void;
  /** Display a message from the assistant. */
  assistantPrint(text: string): void;
  /** Prompt the user for input. */
  promptUser(message: string): Promise<string>;
  /** Confirm an action with the user. */
  confirm(message: string): Promise<boolean>;
}

/** A registered slash command. */
export interface SlashCommand {
  name: string;
  description: string;
  usage?: string;
  handler: (args: string[], ctx: CommandContext) => Promise<void>;
}

// ---------------------------------------------------------------------------
// Registry
// ---------------------------------------------------------------------------

const commands: Map<string, SlashCommand> = new Map();

function register(cmd: SlashCommand): void {
  commands.set(cmd.name, cmd);
}

/** Get a command by name (without leading /). */
export function getCommand(name: string): SlashCommand | undefined {
  return commands.get(name);
}

/** Get all registered commands. */
export function getAllCommands(): SlashCommand[] {
  return Array.from(commands.values());
}

/** Parse and execute a slash command. Returns `false` if the input is not a command. */
export async function executeCommand(input: string, ctx: CommandContext): Promise<boolean> {
  if (!input.startsWith('/')) return false;

  const parts = input.trim().split(/\s+/);
  const commandName = parts[0].slice(1); // strip leading /
  const args = parts.slice(1);

  const cmd = commands.get(commandName);
  if (!cmd) {
    const c = getColors();
    ctx.assistantPrint(c.error(`Unknown command: /${commandName}. Type /help for available commands.`));
    return true;
  }

  try {
    await cmd.handler(args, ctx);
  } catch (err) {
    const c = getColors();
    ctx.assistantPrint(c.error(`Command error: ${String(err)}`));
  }
  return true;
}

// ---------------------------------------------------------------------------
// Built-in commands
// ---------------------------------------------------------------------------

register({
  name: 'help',
  description: 'Show available commands.',
  async handler(_args, ctx) {
    const c = getColors();
    let output = c.primary.bold('\n  Available Commands\n\n');
    for (const cmd of getAllCommands()) {
      output += `  ${c.secondary('/' + cmd.name.padEnd(12))} ${c.muted(cmd.description)}\n`;
    }
    output += '\n  ' + c.muted('Tip: just type your question for normal chat.') + '\n';
    ctx.assistantPrint(output);
  },
});

register({
  name: 'new',
  description: 'Start a new conversation.',
  async handler(_args, ctx) {
    ctx.newConversation();
    ctx.assistantPrint(getColors().success('Started a new conversation.'));
  },
});

register({
  name: 'model',
  description: 'Switch the AI provider and model.',
  usage: '/model [provider] [model]',
  async handler(args, ctx) {
    const c = getColors();
    if (args.length === 0) {
      ctx.assistantPrint(
        `\n  ${c.primary.bold('Current:')} ${c.secondary(ctx.config.provider)} / ${c.secondary(ctx.config.model)}\n\n` +
        `  ${c.muted('Available providers:')}\n` +
        Object.values(AIProviderEnum)
          .map((p) => `    ${c.secondary(p)}`)
          .join('\n') +
        `\n\n  ${c.muted('Usage:')} /model <provider> [model]\n`,
      );
      return;
    }

    const provider = args[0] as AIProviderEnum;
    if (!Object.values(AIProviderEnum).includes(provider)) {
      ctx.assistantPrint(c.error(`Invalid provider: ${provider}`));
      return;
    }

    const model = args[1] || defaultModel(provider);
    const baseUrl = ctx.config.baseUrl || defaultBaseUrl(provider);

    ctx.updateConfig({ provider, model, baseUrl });
    ctx.assistantPrint(c.success(`Switched to ${provider} / ${model}`));
  },
});

register({
  name: 'review',
  description: 'Enter code review mode.',
  async handler(_args, ctx) {
    const prompt = codeReviewPrompt();
    ctx.assistantPrint(getColors().accent('Entering code review mode. Paste your diff or code to review.\n'));
    ctx.conversation.messages.push({ role: 'system', content: prompt, timestamp: Date.now() });
  },
});

register({
  name: 'refactor',
  description: 'Enter refactoring mode.',
  async handler(_args, ctx) {
    const prompt = refactorPrompt();
    ctx.assistantPrint(getColors().accent('Entering refactoring mode. Paste code or describe what to refactor.\n'));
    ctx.conversation.messages.push({ role: 'system', content: prompt, timestamp: Date.now() });
  },
});

register({
  name: 'explain',
  description: 'Explain the provided code.',
  async handler(_args, ctx) {
    const prompt = explainPrompt();
    ctx.assistantPrint(getColors().accent('Entering explanation mode. Paste code to explain.\n'));
    ctx.conversation.messages.push({ role: 'system', content: prompt, timestamp: Date.now() });
  },
});

register({
  name: 'test',
  description: 'Generate tests for the provided code or file.',
  async handler(_args, ctx) {
    const prompt = testGenerationPrompt();
    ctx.assistantPrint(getColors().accent('Entering test generation mode. Paste code or specify a file.\n'));
    ctx.conversation.messages.push({ role: 'system', content: prompt, timestamp: Date.now() });
  },
});

register({
  name: 'commit',
  description: 'Generate a commit message from staged changes.',
  async handler(_args, ctx) {
    const c = getColors();
    const prompt = commitMessagePrompt();
    const provider = createProvider({
      provider: ctx.config.provider,
      model: ctx.config.model,
      apiKey: ctx.config.apiKey,
      baseUrl: ctx.config.baseUrl,
      temperature: ctx.config.temperature,
      maxTokens: ctx.config.maxTokens,
      thinkingEnabled: ctx.config.thinkingEnabled,
    });

    const diff = ctx.toolRegistry.get('gitDiff')?.execute({ staged: true });
    const diffResult = await diff;
    if (!diffResult || !diffResult.success || !diffResult.output) {
      ctx.assistantPrint(c.warning('No staged changes found. Stage your changes first with `git add`.'));
      return;
    }

    const messages: Message[] = [
      { role: 'user', content: `Generate a commit message for this diff:\n\n${diffResult.output}`, timestamp: Date.now() },
    ];

    const result = await provider.generate(messages, prompt);
    ctx.assistantPrint(renderMarkdown(result.content));
  },
});

register({
  name: 'pr',
  description: 'Generate a PR description from recent commits.',
  async handler(_args, ctx) {
    const c = getColors();
    const prompt = prDescriptionPrompt();
    const provider = createProvider({
      provider: ctx.config.provider,
      model: ctx.config.model,
      apiKey: ctx.config.apiKey,
      baseUrl: ctx.config.baseUrl,
      temperature: ctx.config.temperature,
      maxTokens: ctx.config.maxTokens,
      thinkingEnabled: ctx.config.thinkingEnabled,
    });

    const logResult = await ctx.toolRegistry.get('runCommand')?.execute({ command: 'git log --oneline -20' });
    const diffResult = await ctx.toolRegistry.get('gitDiff')?.execute({});

    if (!logResult?.success) {
      ctx.assistantPrint(c.warning('Not in a git repository or no commits found.'));
      return;
    }

    const content = `Generate a PR description.\n\nRecent commits:\n${logResult.output}\n\nUncommitted diff:\n${diffResult?.output ?? '(none)'}`;
    const messages: Message[] = [{ role: 'user', content, timestamp: Date.now() }];

    const result = await provider.generate(messages, prompt);
    ctx.assistantPrint(renderMarkdown(result.content));
  },
});

register({
  name: 'config',
  description: 'Show or edit configuration.',
  usage: '/config [key] [value]',
  async handler(args, ctx) {
    const c = getColors();
    if (args.length === 0) {
      // Show current config (mask apiKey)
      const safe = { ...ctx.config, apiKey: ctx.config.apiKey ? '****' + ctx.config.apiKey.slice(-4) : '(not set)' };
      ctx.assistantPrint(c.primary.bold('\n  Current Configuration\n'));
      for (const [key, value] of Object.entries(safe)) {
        ctx.assistantPrint(`  ${c.secondary(key.padEnd(20))} ${c.muted(String(value))}`);
      }
      ctx.assistantPrint('');
      return;
    }

    const key = args[0];
    const value = args.slice(1).join(' ');
    const validKeys = ['provider', 'model', 'apiKey', 'baseUrl', 'temperature', 'maxTokens', 'theme', 'userName'];

    if (!validKeys.includes(key)) {
      ctx.assistantPrint(c.error(`Unknown config key: ${key}. Valid: ${validKeys.join(', ')}`));
      return;
    }

    const patch: Record<string, unknown> = {};
    if (key === 'temperature' || key === 'maxTokens') {
      patch[key] = Number(value);
    } else if (key === 'thinkingEnabled') {
      patch[key] = value === 'true';
    } else {
      patch[key] = value;
    }

    const errors = validateConfig(patch as PartialConfig);
    if (errors.length > 0) {
      ctx.assistantPrint(c.error(errors.join('\n')));
      return;
    }

    ctx.updateConfig(patch as PartialConfig);
    saveGlobalConfig(patch as PartialConfig);
    ctx.assistantPrint(c.success(`Updated ${key} = ${value}`));
  },
});

register({
  name: 'theme',
  description: 'Change the UI theme.',
  usage: '/theme [dark|light|monokai|dracula]',
  async handler(args, ctx) {
    const c = getColors();
    const validThemes = ['dark', 'light', 'monokai', 'dracula'] as const;

    if (args.length === 0) {
      ctx.assistantPrint(`\n  ${c.primary.bold('Current theme:')} ${getTheme()}\n  ${c.muted('Available:')} ${validThemes.join(', ')}\n`);
      return;
    }

    const theme = args[0] as typeof validThemes[number];
    if (!validThemes.includes(theme)) {
      ctx.assistantPrint(c.error(`Invalid theme: ${theme}. Valid: ${validThemes.join(', ')}`));
      return;
    }

    setTheme(theme);
    ctx.updateConfig({ theme });
    saveGlobalConfig({ theme });
    ctx.assistantPrint(c.success(`Theme changed to ${theme}`));
  },
});

register({
  name: 'skills',
  description: 'List available skills.',
  async handler(_args, ctx) {
    const c = getColors();
    ctx.assistantPrint(c.primary.bold('\n  Available Skills\n'));
    if (ctx.skills.length === 0) {
      ctx.assistantPrint(c.muted('  No skills loaded.\n'));
      ctx.assistantPrint(c.muted('  Create YAML files in .aidev/skills/ or ~/.aidev/skills/\n'));
      return;
    }
    for (const skill of ctx.skills) {
      const trigger = skill.trigger ? c.accent(`Trigger: ${skill.trigger}`) : '';
      ctx.assistantPrint(`  ${c.secondary(skill.name)}  ${c.muted(skill.description)}  ${trigger}`);
    }
    ctx.assistantPrint('');
  },
});

register({
  name: 'plugins',
  description: 'List installed plugins.',
  async handler(_args, ctx) {
    const c = getColors();
    ctx.assistantPrint(c.primary.bold('\n  Installed Plugins\n'));
    ctx.assistantPrint(c.muted('  No plugins installed yet.\n'));
    ctx.assistantPrint(c.muted('  Place plugin directories in ~/.aidev/plugins/\n'));
  },
});

register({
  name: 'history',
  description: 'Show conversation history.',
  async handler(_args, ctx) {
    const c = getColors();
    ctx.assistantPrint(c.primary.bold('\n  Conversation History\n'));
    const userMessages = ctx.conversation.messages.filter((m) => m.role === 'user');
    if (userMessages.length === 0) {
      ctx.assistantPrint(c.muted('  No messages yet.\n'));
      return;
    }
    for (let i = 0; i < userMessages.length; i++) {
      const msg = userMessages[i];
      const preview = msg.content.length > 80 ? msg.content.slice(0, 80) + '...' : msg.content;
      ctx.assistantPrint(`  ${c.muted(`[${i + 1}]`)} ${preview}`);
    }
    ctx.assistantPrint('');
  },
});

register({
  name: 'undo',
  description: 'Undo the last file change (not yet implemented).',
  async handler(_args, ctx) {
    ctx.assistantPrint(getColors().muted('Undo is not yet implemented. Use `git checkout` to revert changes.'));
  },
});

register({
  name: 'export',
  description: 'Export the conversation to a markdown file.',
  async handler(args, ctx) {
    const c = getColors();
    const fs = require('fs') as typeof import('fs');
    const filename = args[0] || `aidev-conversation-${Date.now()}.md`;
    const lines: string[] = [`# aidev Conversation Export\n`];
    lines.push(`**Model:** ${ctx.config.provider} / ${ctx.config.model}\n`);
    lines.push(`**Date:** ${new Date().toISOString()}\n\n---\n`);

    for (const msg of ctx.conversation.messages) {
      if (msg.role === 'system') continue;
      const role = msg.role === 'user' ? '**User**' : '**Assistant**';
      lines.push(`### ${role}\n\n${msg.content}\n\n---\n`);
    }

    try {
      fs.writeFileSync(filename, lines.join('\n'), 'utf-8');
      ctx.assistantPrint(c.success(`Conversation exported to ${filename}`));
    } catch (err) {
      ctx.assistantPrint(c.error(`Failed to export: ${String(err)}`));
    }
  },
});

register({
  name: 'resume',
  description: 'Resume a past conversation from the database.',
  async handler(args, ctx) {
    const c = getColors();

    let sessionId = args[0];
    if (!sessionId) {
      // Show the most recent session
      const sessions = listSessions(5);
      if (sessions.length === 0) {
        ctx.assistantPrint(c.muted('No saved sessions found.'));
        return;
      }
      sessionId = sessions[0].id;
    }

    const session = loadSession(sessionId);
    if (!session) {
      ctx.assistantPrint(c.error(`Session not found: ${sessionId}`));
      return;
    }

    // Replace current conversation
    ctx.conversation.id = session.id;
    ctx.conversation.title = session.title;
    ctx.conversation.messages = session.messages;
    ctx.conversation.provider = session.provider;
    ctx.conversation.model = session.model;
    ctx.assistantPrint(c.success(`Resumed session: ${session.title} (${session.messages.length} messages)`));
  },
});

register({
  name: 'sessions',
  description: 'List recent saved sessions.',
  async handler(args, ctx) {
    const c = getColors();
    const limit = args[0] ? parseInt(args[0], 10) : 10;
    const sessions = listSessions(limit);

    if (sessions.length === 0) {
      ctx.assistantPrint(c.muted('No saved sessions found.'));
      return;
    }

    ctx.assistantPrint(c.primary.bold('\n  Recent Sessions\n'));
    for (const s of sessions) {
      ctx.assistantPrint(`  ${c.secondary(s.id.slice(0, 16))}  ${c.muted(s.title)}  ${c.muted(`(${s.message_count} msgs, ${s.updated_at})`)}`);
    }
    ctx.assistantPrint('');
  },
});

register({
  name: 'clear',
  description: 'Clear the terminal screen.',
  async handler(_args, ctx) {
    process.stdout.write('\x1Bc');
    // Re-print banner/status after clear
    ctx.assistantPrint('');
  },
});

register({
  name: 'exit',
  description: 'Exit aidev.',
  async handler() {
    process.exit(0);
  },
});
