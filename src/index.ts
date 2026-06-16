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
import * as fs from 'fs';
import * as path from 'path';
import chalk from 'chalk';

import { AppConfig, Conversation, Message, Skill, ToolCall, ToolDefinition } from './types';
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
  renderDiff,
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
import { saveSession } from './storage/session';

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

// Read version from package.json (avoids hardcoding)
const VERSION: string = (() => {
  try {
    const pkgPath = path.join(__dirname, '..', 'package.json');
    return JSON.parse(fs.readFileSync(pkgPath, 'utf-8')).version ?? '0.0.0';
  } catch {
    return '0.0.0';
  }
})();

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
    skills,
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

/** Maximum estimated tokens before triggering compaction. */
const COMPACTION_THRESHOLD = 120_000;

/**
 * Lightweight context compaction: when the conversation exceeds the token
 * threshold, replace the older half with a summary placeholder.
 * (Inspired by DeepSeek-Reasonix's compaction system — summaries accumulate,
 * never re-summarize, and tool results are pruned first.)
 */
function compactIfNeeded(): void {
  const totalTokens = estimateConversationTokens(conversation.messages);
  if (totalTokens < COMPACTION_THRESHOLD) return;

  const c = getColors();
  console.log(c.muted(`\n  [Compacting: ~${Math.round(totalTokens)} tokens estimated]`));

  // Prune first to remove re-derivable data
  conversation.messages = pruneToolResults(conversation.messages);

  // Find a good boundary: keep system messages + first user turn + recent tail
  const recentCount = Math.min(6, conversation.messages.length);
  const keepRecent = conversation.messages.slice(-recentCount);
  const foldable = conversation.messages.slice(0, -recentCount);

  if (foldable.length === 0) return;

  // Don't re-compact already-compacted sections
  if (foldable.every((m) => m.content.startsWith('[Context Summary'))) return;

  // Build a summary placeholder
  const userTurns = foldable.filter((m) => m.role === 'user').length;
  const assistantTurns = foldable.filter((m) => m.role === 'assistant').length;
  const toolCalls = foldable.reduce((acc, m) => acc + (m.toolCalls?.length ?? 0), 0);
  const topics = foldable.filter((m) => m.role === 'user').map((m) => m.content.slice(0, 60)).join('; ');

  const summary: Message = {
    role: 'system',
    content: `[Context Summary: ${userTurns} user turns, ${assistantTurns} assistant responses, ${toolCalls} tool calls. Topics discussed: ${topics.slice(0, 300)}]`,
    timestamp: Date.now(),
  };

  // Keep system messages from the start, then summary, then recent tail
  const systemMessages = foldable.filter((m) => m.role === 'system' && !m.content.startsWith('[Context Summary'));
  conversation.messages = [...systemMessages.slice(0, 1), summary, ...keepRecent];

  const newTokens = estimateConversationTokens(conversation.messages);
  console.log(c.muted(`  [Compacted to ~${Math.round(newTokens)} tokens]`));
}

// ---------------------------------------------------------------------------
// Process a single user message (chat + AI interaction with tool calling)
// ---------------------------------------------------------------------------

/** Maximum number of tool call rounds to prevent doom loops */
const MAX_TOOL_ROUNDS = 10;

/** Storm breaker threshold: same (tool, error) signature N times → abort. */
const STORM_BREAK_THRESHOLD = 3;

/**
 * Detect if the model is stuck in a death spiral.
 * Keyed on (toolName, error) rather than (toolName, args) because the model
 * keeps cosmetically reworking arguments while the failure stays identical.
 * (Inspired by DeepSeek-Reasonix's storm breaker.)
 */
class StormBreaker {
  private signatures = new Map<string, number>();

  /** Record a failed tool call and return true if the storm threshold is breached. */
  record(toolName: string, error: string | undefined): boolean {
    if (!error) return false;
    const sig = `${toolName}:${error.slice(0, 120)}`;
    const count = (this.signatures.get(sig) ?? 0) + 1;
    this.signatures.set(sig, count);
    return count >= STORM_BREAK_THRESHOLD;
  }

  reset(): void {
    this.signatures.clear();
  }
}

/**
 * Token estimation: calibrated from real API usage instead of a fixed heuristic.
 * (Inspired by DeepSeek-Reasonix's calibrated tokPerChar.)
 */
let tokPerChar = 0.25; // initial guess (~4 chars/token for English)

function estimateTokens(text: string): number {
  return Math.ceil(text.length * tokPerChar);
}

function calibrateFromUsage(realUsage: { promptTokens: number; completionTokens: number }, sampleText: string): void {
  if (sampleText.length > 0) {
    const measured = sampleText.length / realUsage.completionTokens;
    // Exponential moving average to smoothly adapt
    tokPerChar = tokPerChar * 0.7 + (1 / measured) * 0.3;
  }
}

/** Display token usage and estimated cost. */
function displayTokenUsage(result: { usage: { promptTokens: number; completionTokens: number; totalTokens: number }; estimatedCost: number }): void {
  if (!config.showTokenCounts || result.usage.totalTokens <= 0) return;
  const c = getColors();
  const costStr = result.estimatedCost > 0 ? ` (~$${result.estimatedCost.toFixed(4)})` : '';
  console.log(
    c.muted(
      `\n  Tokens: ${result.usage.promptTokens} in / ${result.usage.completionTokens} out / ${result.usage.totalTokens} total${costStr}`,
    ),
  );
}

/**
 * Estimate total tokens in the conversation.
 * Uses the calibrated tokPerChar for a quick estimate without API calls.
 */
function estimateConversationTokens(messages: Message[]): number {
  let total = 0;
  for (const msg of messages) {
    total += estimateTokens(msg.content || '');
    if (msg.toolCalls) {
      for (const tc of msg.toolCalls) {
        total += estimateTokens(JSON.stringify(tc.arguments));
      }
    }
  }
  return total;
}

/**
 * Prune re-derivable tool results to save context space.
 * Files can be re-read, search results re-obtained.
 * Replaces verbose outputs with compact placeholders.
 * (Inspired by DeepSeek-Reasonix's tool result pruning.)
 */
function pruneToolResults(messages: Message[]): Message[] {
  const REPRUNABLE = new Set(['readFile', 'searchFiles', 'listDir', 'gitStatus', 'gitDiff']);
  const MAX_TOOL_OUTPUT = 500; // chars to keep for pruned results

  // Only prune messages older than the last 3 rounds (keep recent context intact)
  const cutoff = Math.max(0, messages.length - 8);

  return messages.map((msg, i) => {
    if (i >= cutoff) return msg;
    if (msg.role !== 'tool' || !msg.toolName || !REPRUNABLE.has(msg.toolName)) return msg;

    // Already pruned
    if (msg.content.startsWith('[Pruned:')) return msg;

    const originalLen = msg.content.length;
    if (originalLen <= MAX_TOOL_OUTPUT) return msg;

    return {
      ...msg,
      content: `[Pruned: ${msg.toolName} output (${originalLen} chars). Re-run the tool to get fresh data.]`,
    };
  });
}

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

  // Auto-title: use first user message as session title
  if (conversation.title === 'New Conversation' && conversation.messages.filter((m) => m.role === 'user').length === 1) {
    conversation.title = userInput.length > 60 ? userInput.slice(0, 57) + '...' : userInput;
  }

  // Build system prompt
  const systemPrompt = codingAssistantPrompt(projectContext);

  // Auto-compact if conversation is getting too long
  compactIfNeeded();

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

  // Get tool definitions from registry
  const toolDefs: ToolDefinition[] = toolRegistry.getDefinitions();

  // Ask-user callback for permission system
  const askUser = async (toolName: string): Promise<boolean> => {
    return confirmAction(`Allow tool "${toolName}" to run?`);
  };

  // Storm breaker: detect death-spiral loops by (tool, error) signature
  const stormBreaker = new StormBreaker();

  // Tool call loop: keep calling AI until it returns text (no more tool calls)
  for (let round = 0; round < MAX_TOOL_ROUNDS; round++) {
    // Prune old tool results before sending to AI (saves context space)
    const prunedMessages = pruneToolResults(conversation.messages);

    let result: Awaited<ReturnType<typeof provider.generate>>;

    if (config.stream) {
      // Streaming mode
      const writer = new StreamWriter();
      try {
        result = await provider.stream(
          prunedMessages,
          (chunk) => {
            if (chunk.thinking && config.thinkingEnabled) {
              process.stdout.write(c.thinking(chunk.thinking));
            }
            if (chunk.content) {
              writer.write(chunk.content);
            }
          },
          systemPrompt,
          toolDefs.length > 0 ? toolDefs : undefined,
        );
        writer.end();
        // Calibrate token estimation from real usage
        calibrateFromUsage(result.usage, result.content);
      } catch (err) {
        writer.end();
        console.error(c.error(`\nError: ${String(err)}`));
        return;
      }
    } else {
      // Non-streaming mode
      const spin = spinner('Thinking...');
      try {
        result = await provider.generate(
          prunedMessages,
          systemPrompt,
          toolDefs.length > 0 ? toolDefs : undefined,
        );
        spin.stop();
        // Calibrate token estimation from real usage
        calibrateFromUsage(result.usage, result.content);

        if (result.thinking && config.thinkingEnabled) {
          console.log(c.thinking(`[Thinking]\n${result.thinking}\n`));
        }

        if (result.content) {
          console.log(renderMarkdown(result.content));
        }
      } catch (err) {
        spin.stop();
        console.error(c.error(`\nError: ${String(err)}`));
        return;
      }
    }

    // Store assistant response
    const assistantMessage: Message = {
      role: 'assistant',
      content: result.content,
      timestamp: Date.now(),
      ...(result.toolCalls && result.toolCalls.length > 0 ? { toolCalls: result.toolCalls } : {}),
    };
    conversation.messages.push(assistantMessage);

    // Show token usage
    displayTokenUsage(result);

    // If no tool calls, we're done
    if (!result.toolCalls || result.toolCalls.length === 0) {
      // Save conversation to database
      try { saveSession(conversation); } catch { /* best-effort save */ }
      return;
    }

    // Execute tool calls
    console.log(c.accent(`\n  [Tool Calls: ${result.toolCalls.map((tc) => tc.name).join(', ')}]\n`));

    // Parallel dispatch for all-readOnly batches (inspired by Reasonix's parallel execution)
    const allReadOnly = result.toolCalls.every((tc) => {
      const tool = toolRegistry.get(tc.name);
      return tool?.readOnly === true;
    });

    interface ToolCallResult {
      toolCall: ToolCall;
      toolResult: Awaited<ReturnType<ToolRegistry['execute']>>;
    }

    let toolCallResults: ToolCallResult[];

    if (allReadOnly && result.toolCalls.length > 1) {
      // All read-only → execute in parallel
      toolCallResults = await Promise.all(
        result.toolCalls.map(async (tc) => ({
          toolCall: tc,
          toolResult: await toolRegistry.execute(tc.name, tc.arguments, config.permissions, config.defaultPermission),
        })),
      );
    } else {
      // Has write tools → execute sequentially
      toolCallResults = [];
      for (const tc of result.toolCalls) {
        toolCallResults.push({
          toolCall: tc,
          toolResult: await toolRegistry.execute(tc.name, tc.arguments, config.permissions, config.defaultPermission, askUser),
        });
      }
    }

    for (const { toolCall, toolResult } of toolCallResults) {

      // Show tool result summary
      if (toolResult.success) {
        if (toolCall.name === 'editFile' && toolResult.output.includes('Diff:')) {
          // Edit preview: show diff prominently
          console.log(c.success(`  ✓ ${toolCall.name} applied:`));
          const diffPart = toolResult.output.split('Diff:\n')[1] ?? toolResult.output;
          console.log(renderDiff(diffPart));
        } else {
          const preview = toolResult.output.length > 200
            ? toolResult.output.slice(0, 200) + '...'
            : toolResult.output;
          console.log(c.success(`  ✓ ${toolCall.name}: `) + c.muted(preview.replace(/\n/g, ' ')));
        }
      } else {
        console.log(c.error(`  ✗ ${toolCall.name}: ${toolResult.error}`));
        // Storm breaker: check for death-spiral loops
        if (stormBreaker.record(toolCall.name, toolResult.error)) {
          console.log(c.error(`\n  ⚠ Storm breaker: same tool/error combination repeated ${STORM_BREAK_THRESHOLD}+ times. Stopping loop.`));
          conversation.messages.push({
            role: 'assistant',
            content: `[System: Tool call loop detected — same error "${toolResult.error?.slice(0, 80)}" repeating. Please try a different approach.]`,
            timestamp: Date.now(),
          });
          saveSession(conversation);
          return;
        }
      }

      // Add tool result to conversation
      conversation.messages.push({
        role: 'tool',
        content: toolResult.success ? toolResult.output : `Error: ${toolResult.error}`,
        timestamp: Date.now(),
        toolCallId: toolCall.id,
        toolName: toolCall.name,
      });
    }

    console.log(''); // separator before next round
    // Loop continues — AI will see the tool results and may call more tools
  }

  console.log(c.warning(`\n  [Reached maximum tool call rounds (${MAX_TOOL_ROUNDS})]`));
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
      const toolDefs: ToolDefinition[] = toolRegistry.getDefinitions();

      // Tool call loop for single-question mode
      for (let round = 0; round < MAX_TOOL_ROUNDS; round++) {
        let result: Awaited<ReturnType<typeof provider.generate>>;

        if (config.stream) {
          // Streaming mode
          const writer = new StreamWriter();
          const spin = spinner('Thinking...');
          try {
            result = await provider.stream(
              messages,
              (chunk) => {
                spin.stop();
                if (chunk.thinking && config.thinkingEnabled) {
                  process.stdout.write(getColors().thinking(chunk.thinking));
                }
                if (chunk.content) writer.write(chunk.content);
              },
              systemPrompt,
              toolDefs.length > 0 ? toolDefs : undefined,
            );
            writer.end();
          } catch (err) {
            writer.end();
            console.error(chalk.red(`Error: ${String(err)}`));
            process.exit(1);
          }
        } else {
          // Non-streaming mode
          const spin = spinner('Thinking...');
          try {
            result = await provider.generate(messages, systemPrompt, toolDefs.length > 0 ? toolDefs : undefined);
            spin.stop();
            if (result.content) console.log(renderMarkdown(result.content));
          } catch (err) {
            spin.stop();
            console.error(chalk.red(`Error: ${String(err)}`));
            process.exit(1);
          }
        }

        if (!result!.toolCalls || result!.toolCalls.length === 0) {
          if (config.showTokenCounts && result!.usage.totalTokens > 0) {
            const c = getColors();
            console.log(c.muted(`\n  Tokens: ${result!.usage.totalTokens} total`));
          }
          process.exit(0);
        }

        // Execute tool calls
        messages.push({ role: 'assistant', content: result!.content, timestamp: Date.now(), toolCalls: result!.toolCalls });
        for (const toolCall of result!.toolCalls) {
          const toolResult = await toolRegistry.execute(toolCall.name, toolCall.arguments, config.permissions, config.defaultPermission);
          messages.push({ role: 'tool', content: toolResult.success ? toolResult.output : `Error: ${toolResult.error}`, timestamp: Date.now(), toolCallId: toolCall.id, toolName: toolCall.name });
        }
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
