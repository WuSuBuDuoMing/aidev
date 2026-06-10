#!/usr/bin/env node

/**
 * aidev CLI — Main Entry Point
 *
 * Parses CLI arguments, initializes all subsystems, and runs the interactive
 * REPL loop. Handles graceful shutdown on SIGINT/SIGTERM.
 *
 * @module index
 */

import { Command } from 'commander';
import * as readline from 'readline';
import chalk from 'chalk';

import { AppConfig, Conversation, Message, Skill } from './types';
import { loadConfig, saveGlobalConfig } from './config';
import { createProvider, defaultModel, defaultBaseUrl } from './ai/provider';
import { codingAssistantPrompt } from './ai/prompts';
import { analyzeProject, contextSummary } from './context/analyzer';
import { ToolRegistry, createBuiltinRegistry } from './tools';
import {
  setTheme,
  printBanner,
  statusBar,
  horizontalRule,
  renderMarkdown,
  StreamWriter,
  spinner,
  getColors,
  confirmAction,
} from './ui/terminal';
import {
  CommandContext,
  executeCommand,
  getAllCommands,
} from './commands';
import { loadAllSkills, resolveSkill } from './skills';

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

const VERSION = '1.0.0';

// ---------------------------------------------------------------------------
// REPL state
// ---------------------------------------------------------------------------

let config: AppConfig;
let conversation: Conversation;
let toolRegistry: ToolRegistry;
let projectContext: ReturnType<typeof analyzeProject>;
let skills: Skill[];

function newConversation(): void {
  conversation = {
    id: `conv-${Date.now()}`,
    title: 'New Conversation',
    messages: [],
    createdAt: Date.now(),
    updatedAt: Date.now(),
    model: config.model,
    provider: config.provider,
  };
}

// ---------------------------------------------------------------------------
// Build the CommandContext object for the commands module
// ---------------------------------------------------------------------------

function buildCommandContext(): CommandContext {
  return {
    config,
    conversation,
    projectContext,
    toolRegistry,
    updateConfig(patch) {
      Object.assign(config, patch);
    },
    newConversation() {
      newConversation();
    },
    refreshContext() {
      projectContext = analyzeProject();
    },
    assistantPrint(text: string) {
      console.log(text);
    },
    promptUser(message: string): Promise<string> {
      return new Promise((resolve) => {
        const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
        rl.question(getColors().primary(message), (answer) => {
          rl.close();
          resolve(answer);
        });
      });
    },
    confirm(message: string): Promise<boolean> {
      return confirmAction(message);
    },
  };
}

// ---------------------------------------------------------------------------
// Process a single user message (chat + AI interaction)
// ---------------------------------------------------------------------------

async function processMessage(userInput: string): Promise<void> {
  const c = getColors();
  const ctx = buildCommandContext();

  // Check slash commands first
  if (userInput.startsWith('/')) {
    // Check if it's a skill trigger
    const skill = resolveSkill(skills, userInput);
    if (skill) {
      // Inject the skill prompt into the conversation
      conversation.messages.push({
        role: 'system',
        content: skill.prompt,
        timestamp: Date.now(),
      });
      console.log(c.accent(`[Skill: ${skill.name}] ${skill.description}\n`));
      return;
    }

    const handled = await executeCommand(userInput, ctx);
    if (handled) return;
  }

  // Normal chat message
  const userMessage: Message = {
    role: 'user',
    content: userInput,
    timestamp: Date.now(),
  };
  conversation.messages.push(userMessage);
  conversation.updatedAt = Date.now();

  // Build system prompt
  const systemPrompt = codingAssistantPrompt(projectContext);

  // Create provider
  const provider = createProvider({
    provider: config.provider,
    model: config.model,
    apiKey: config.apiKey,
    baseUrl: config.baseUrl,
    temperature: config.temperature,
    maxTokens: config.maxTokens,
    thinkingEnabled: config.thinkingEnabled,
  });

  if (config.stream) {
    // Streaming mode
    const writer = new StreamWriter();
    try {
      const result = await provider.stream(
        conversation.messages,
        (chunk) => {
          if (chunk.thinking && config.thinkingEnabled) {
            // Show thinking in muted italic
            process.stdout.write(c.thinking(chunk.thinking));
          }
          if (chunk.content) {
            writer.write(chunk.content);
          }
        },
        systemPrompt,
      );
      writer.end();

      // Store assistant response
      conversation.messages.push({
        role: 'assistant',
        content: result.content,
        timestamp: Date.now(),
      });

      // Show token usage
      if (config.showTokenCounts && result.usage.totalTokens > 0) {
        const costStr = result.estimatedCost > 0
          ? ` (~$${result.estimatedCost.toFixed(4)})`
          : '';
        console.log(
          c.muted(
            `\n  Tokens: ${result.usage.promptTokens} in / ${result.usage.completionTokens} out / ${result.usage.totalTokens} total${costStr}`,
          ),
        );
      }
    } catch (err) {
      writer.end();
      console.error(c.error(`\nError: ${String(err)}`));
    }
  } else {
    // Non-streaming mode
    const spin = spinner('Thinking...');
    try {
      const result = await provider.generate(conversation.messages, systemPrompt);
      spin.stop();

      // Display thinking if available
      if (result.thinking && config.thinkingEnabled) {
        console.log(c.thinking(`[Thinking]\n${result.thinking}\n`));
      }

      console.log(renderMarkdown(result.content));

      conversation.messages.push({
        role: 'assistant',
        content: result.content,
        timestamp: Date.now(),
      });

      if (config.showTokenCounts && result.usage.totalTokens > 0) {
        const costStr = result.estimatedCost > 0
          ? ` (~$${result.estimatedCost.toFixed(4)})`
          : '';
        console.log(
          c.muted(
            `\n  Tokens: ${result.usage.promptTokens} in / ${result.usage.completionTokens} out / ${result.usage.totalTokens} total${costStr}`,
          ),
        );
      }
    } catch (err) {
      spin.stop();
      console.error(c.error(`\nError: ${String(err)}`));
    }
  }
}

// ---------------------------------------------------------------------------
// REPL loop
// ---------------------------------------------------------------------------

async function repl(): Promise<void> {
  const c = getColors();
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: c.primary('aidev> '),
  });

  console.log(statusBar(config.provider, config.model, config.theme));
  console.log(c.muted('  Type /help for commands, or just start chatting.\n'));

  rl.prompt();

  rl.on('line', async (line: string) => {
    const input = line.trim();
    if (!input) {
      rl.prompt();
      return;
    }

    try {
      await processMessage(input);
    } catch (err) {
      console.error(c.error(`Error: ${String(err)}`));
    }

    console.log(''); // blank line separator
    rl.prompt();
  });

  rl.on('close', () => {
    gracefulShutdown();
  });
}

// ---------------------------------------------------------------------------
// Graceful shutdown
// ---------------------------------------------------------------------------

function gracefulShutdown(): void {
  const c = getColors();
  console.log(c.muted('\nGoodbye! Happy coding.'));
  process.exit(0);
}

process.on('SIGINT', () => gracefulShutdown());
process.on('SIGTERM', () => gracefulShutdown());

// ---------------------------------------------------------------------------
// CLI argument parsing
// ---------------------------------------------------------------------------

const program = new Command();

program
  .name('aidev')
  .description('AI-powered terminal coding assistant — supports Claude, GPT, DeepSeek, Gemini and any OpenAI-compatible API')
  .version(VERSION);

program
  .option('-p, --provider <provider>', 'AI provider to use')
  .option('-m, --model <model>', 'Model to use')
  .option('-k, --api-key <key>', 'API key')
  .option('-b, --base-url <url>', 'Base URL for the API')
  .option('-t, --temperature <temp>', 'Temperature (0-2)', parseFloat)
  .option('--max-tokens <n>', 'Maximum tokens to generate', parseInt)
  .option('--thinking', 'Enable thinking/reasoning mode')
  .option('--theme <theme>', 'UI theme (dark, light, monokai, dracula)')
  .option('--no-stream', 'Disable streaming output')
  .option('-q, --question <question>', 'Ask a single question and exit')
  .option('--status', 'Show current configuration and exit')
  .action(async (opts) => {
    // Load config
    try {
      config = loadConfig();
    } catch (err) {
      console.error(chalk.red(`Configuration error: ${String(err)}`));
      process.exit(1);
    }

    // Apply CLI overrides
    if (opts.provider) config.provider = opts.provider;
    if (opts.model) config.model = opts.model;
    if (opts.apiKey) config.apiKey = opts.apiKey;
    if (opts.baseUrl) config.baseUrl = opts.baseUrl;
    if (opts.temperature !== undefined) config.temperature = opts.temperature;
    if (opts.maxTokens !== undefined) config.maxTokens = opts.maxTokens;
    if (opts.thinking) config.thinkingEnabled = true;
    if (opts.theme) config.theme = opts.theme;
    if (opts.stream === false) config.stream = false;

    // Initialize subsystems
    setTheme(config.theme);
    toolRegistry = createBuiltinRegistry();
    projectContext = analyzeProject();
    skills = loadAllSkills();

    // --status: print config and exit
    if (opts.status) {
      const c = getColors();
      console.log(c.primary.bold('\n  aidev Status\n'));
      console.log(`  Provider:     ${c.secondary(config.provider)}`);
      console.log(`  Model:        ${c.secondary(config.model)}`);
      console.log(`  API Key:      ${config.apiKey ? c.success('set') : c.warning('not set')}`);
      console.log(`  Base URL:     ${config.baseUrl || c.muted('(default)')}`);
      console.log(`  Temperature:  ${config.temperature}`);
      console.log(`  Max Tokens:   ${config.maxTokens}`);
      console.log(`  Streaming:    ${config.stream}`);
      console.log(`  Theme:        ${config.theme}`);
      console.log(`  Thinking:     ${config.thinkingEnabled}`);
      console.log(`  Project:      ${projectContext.language} / ${projectContext.framework}`);
      console.log(`  Git Branch:   ${projectContext.gitInfo?.branch ?? 'N/A'}`);
      console.log(`  Skills:       ${skills.length} loaded`);
      console.log(`  Tools:        ${toolRegistry.getAll().length} registered`);
      console.log('');
      process.exit(0);
    }

    // -q: single question mode
    if (opts.question) {
      const provider = createProvider({
        provider: config.provider,
        model: config.model,
        apiKey: config.apiKey,
        baseUrl: config.baseUrl,
        temperature: config.temperature,
        maxTokens: config.maxTokens,
        thinkingEnabled: config.thinkingEnabled,
      });

      const systemPrompt = codingAssistantPrompt(projectContext);
      const messages: Message[] = [
        { role: 'user', content: opts.question, timestamp: Date.now() },
      ];

      const spin = spinner('Thinking...');
      try {
        const result = await provider.generate(messages, systemPrompt);
        spin.stop();
        console.log(renderMarkdown(result.content));
        if (config.showTokenCounts && result.usage.totalTokens > 0) {
          const c = getColors();
          console.log(c.muted(`\n  Tokens: ${result.usage.totalTokens} total`));
        }
      } catch (err) {
        spin.stop();
        console.error(chalk.red(`Error: ${String(err)}`));
        process.exit(1);
      }
      process.exit(0);
    }

    // Check API key
    if (!config.apiKey && config.provider !== 'ollama') {
      const c = getColors();
      console.log(c.warning('\n  No API key configured.'));
      console.log(c.muted('  Set AIDEV_API_KEY, or configure via /config apiKey <key>\n'));
    }

    // Start interactive mode
    newConversation();
    printBanner();
    const c = getColors();
    console.log(c.muted(`  Project: ${projectContext.language} / ${projectContext.framework}`));
    if (projectContext.gitInfo) {
      console.log(c.muted(`  Branch:  ${projectContext.gitInfo.branch}`));
    }
    console.log('');

    await repl();
  });

// Parse and run
program.parse(process.argv);
