/**
 * aidev CLI — Terminal UI
 *
 * Handles all terminal rendering: markdown with syntax highlighting, streaming
 * output, progress spinners, diff display, theming, and user input.
 *
 * @module ui/terminal
 */

import * as readline from 'readline';
import chalk from 'chalk';
import { highlight } from 'cli-highlight';

// ---------------------------------------------------------------------------
// Theme
// ---------------------------------------------------------------------------

/** Available colour themes. */
export type ThemeName = 'dark' | 'light' | 'monokai' | 'dracula';

interface ThemeColors {
  primary: typeof chalk;
  secondary: typeof chalk;
  accent: typeof chalk;
  success: typeof chalk;
  warning: typeof chalk;
  error: typeof chalk;
  muted: typeof chalk;
  code: typeof chalk;
  thinking: typeof chalk;
}

const THEMES: Record<ThemeName, ThemeColors> = {
  dark: {
    primary: chalk.cyan,
    secondary: chalk.white,
    accent: chalk.magenta,
    success: chalk.green,
    warning: chalk.yellow,
    error: chalk.red,
    muted: chalk.gray,
    code: chalk.bgBlack.white,
    thinking: chalk.italic.gray,
  },
  light: {
    primary: chalk.blue,
    secondary: chalk.black,
    accent: chalk.magenta,
    success: chalk.green,
    warning: chalk.hex('#b45309'),
    error: chalk.red,
    muted: chalk.gray,
    code: chalk.bgWhite.black,
    thinking: chalk.italic.gray,
  },
  monokai: {
    primary: chalk.hex('#66d9ef'),
    secondary: chalk.hex('#f8f8f2'),
    accent: chalk.hex('#f92672'),
    success: chalk.hex('#a6e22e'),
    warning: chalk.hex('#e6db74'),
    error: chalk.hex('#f92672'),
    muted: chalk.hex('#75715e'),
    code: chalk.bgHex('#272822').hex('#f8f8f2'),
    thinking: chalk.italic.hex('#75715e'),
  },
  dracula: {
    primary: chalk.hex('#bd93f9'),
    secondary: chalk.hex('#f8f8f2'),
    accent: chalk.hex('#ff79c6'),
    success: chalk.hex('#50fa7b'),
    warning: chalk.hex('#f1fa8c'),
    error: chalk.hex('#ff5555'),
    muted: chalk.hex('#6272a4'),
    code: chalk.bgHex('#282a36').hex('#f8f8f2'),
    thinking: chalk.italic.hex('#6272a4'),
  },
};

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

let currentTheme: ThemeName = 'dark';
let currentColors: ThemeColors = THEMES.dark;

// ---------------------------------------------------------------------------
// Public API — Theming
// ---------------------------------------------------------------------------

/** Set the active colour theme. */
export function setTheme(name: ThemeName): void {
  currentTheme = name;
  currentColors = THEMES[name];
}

/** Get the current theme name. */
export function getTheme(): ThemeName {
  return currentTheme;
}

/** Get the current theme colours (for use in other modules). */
export function getColors(): ThemeColors {
  return currentColors;
}

// ---------------------------------------------------------------------------
// Markdown rendering with syntax highlighting
// ---------------------------------------------------------------------------

/** Naively render markdown to the terminal with syntax-highlighted code blocks. */
export function renderMarkdown(text: string): string {
  const lines = text.split('\n');
  const output: string[] = [];
  let inCodeBlock = false;
  let codeLang = '';
  let codeLines: string[] = [];

  for (const line of lines) {
    // Fenced code block toggle
    if (line.trimStart().startsWith('```')) {
      if (!inCodeBlock) {
        inCodeBlock = true;
        codeLang = line.trimStart().slice(3).trim();
        codeLines = [];
      } else {
        // Close code block
        const code = codeLines.join('\n');
        try {
          const highlighted = highlight(code, { language: codeLang || undefined, ignoreIllegals: true });
          output.push(currentColors.code('┌' + '─'.repeat(60)));
          for (const hl of highlighted.split('\n')) {
            output.push(currentColors.code('│ ') + hl);
          }
          output.push(currentColors.code('└' + '─'.repeat(60)));
        } catch {
          output.push(currentColors.code(code));
        }
        inCodeBlock = false;
        codeLang = '';
        codeLines = [];
      }
      continue;
    }

    if (inCodeBlock) {
      codeLines.push(line);
      continue;
    }

    // Inline rendering
    let rendered = line;

    // Headers
    if (/^#{1,6}\s/.test(rendered)) {
      const level = rendered.match(/^(#{1,6})/)?.[1].length ?? 1;
      rendered = rendered.replace(/^#{1,6}\s+/, '');
      if (level === 1) rendered = currentColors.primary.bold.underline(rendered);
      else if (level === 2) rendered = currentColors.primary.bold(rendered);
      else rendered = currentColors.accent.bold(rendered);
      output.push(rendered);
      continue;
    }

    // Bold + italic
    rendered = rendered.replace(/\*\*\*(.+?)\*\*\*/g, (_, m) => chalk.bold.italic(m));
    rendered = rendered.replace(/\*\*(.+?)\*\*/g, (_, m) => chalk.bold(m));
    rendered = rendered.replace(/\*(.+?)\*/g, (_, m) => chalk.italic(m));

    // Inline code
    rendered = rendered.replace(/`([^`]+)`/g, (_, m) => currentColors.code(` ${m} `));

    // Unordered list
    rendered = rendered.replace(/^(\s*)[-*+]\s/, `$1${currentColors.accent('  ')}• `);

    // Ordered list
    rendered = rendered.replace(/^(\s*)(\d+)\.\s/, `$1${currentColors.accent('  $2.')}`);

    output.push(rendered);
  }

  return output.join('\n');
}

// ---------------------------------------------------------------------------
// Streaming output
// ---------------------------------------------------------------------------

/** Create a streaming writer that prints incremental chunks to stdout. */
export class StreamWriter {
  private buffer = '';
  private lineBuffer = '';

  /** Write a chunk of streamed text. Handles partial lines. */
  write(chunk: string): void {
    this.buffer += chunk;

    // Flush complete lines
    const lines = this.buffer.split('\n');
    this.buffer = lines.pop() ?? '';

    for (const line of lines) {
      this.lineBuffer += line;
      this.flushLine();
    }
  }

  /** Signal that streaming is complete — flush remaining buffer. */
  end(): void {
    if (this.buffer) {
      this.lineBuffer += this.buffer;
    }
    if (this.lineBuffer) {
      this.flushLine();
    }
    this.buffer = '';
    this.lineBuffer = '';
    process.stdout.write('\n');
  }

  private flushLine(): void {
    // Simple inline rendering (full markdown is done at end)
    let line = this.lineBuffer;
    line = line.replace(/`([^`]+)`/g, (_, m) => currentColors.code(` ${m} `));
    line = line.replace(/\*\*(.+?)\*\*/g, (_, m) => chalk.bold(m));
    line = line.replace(/^#{1,6}\s+(.*)/, (_, m) => currentColors.primary.bold(m));
    process.stdout.write(line + '\n');
    this.lineBuffer = '';
  }
}

// ---------------------------------------------------------------------------
// Progress spinner
// ---------------------------------------------------------------------------

/** Display a spinner with a message. Returns a stop function. */
export function spinner(text: string): { stop: (finalText?: string) => void } {
  const frames = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
  let i = 0;
  let running = true;

  const timer = setInterval(() => {
    if (!running) return;
    process.stdout.write(`\r${currentColors.primary(frames[i++ % frames.length])} ${text}`);
  }, 80);

  return {
    stop(finalText?: string) {
      running = false;
      clearInterval(timer);
      process.stdout.write('\r' + ' '.repeat(text.length + 4) + '\r');
      if (finalText) {
        process.stdout.write(`${currentColors.success('✓')} ${finalText}\n`);
      }
    },
  };
}

// ---------------------------------------------------------------------------
// Diff display
// ---------------------------------------------------------------------------

/** Render a unified diff string with colour. */
export function renderDiff(diffText: string): string {
  return diffText
    .split('\n')
    .map((line) => {
      if (line.startsWith('+++') || line.startsWith('---')) return chalk.bold(line);
      if (line.startsWith('+')) return currentColors.success(line);
      if (line.startsWith('-')) return currentColors.error(line);
      if (line.startsWith('@@')) return currentColors.accent(line);
      return line;
    })
    .join('\n');
}

// ---------------------------------------------------------------------------
// Status bar
// ---------------------------------------------------------------------------

/** Render a one-line status bar. */
export function statusBar(provider: string, model: string, theme: ThemeName): string {
  const left = ` ${currentColors.primary(provider)} ${currentColors.secondary('/')}`;
  const right = `${currentColors.muted(model)} ${currentColors.secondary('/')} ${currentColors.muted(theme)} `;
  const width = process.stdout.columns || 80;
  const padding = Math.max(0, width - left.length - right.length - 2);
  return `${currentColors.accent('─'.repeat(width))}\n${left}${' '.repeat(padding)}${right}\n${currentColors.accent('─'.repeat(width))}`;
}

// ---------------------------------------------------------------------------
// Input
// ---------------------------------------------------------------------------

/** Prompt the user for a single line of input. */
export function promptUser(prompt: string): Promise<string> {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(currentColors.primary(prompt), (answer) => {
      rl.close();
      resolve(answer);
    });
  });
}

/** Prompt with multi-line support: submit on Enter, Shift+Enter for newline. */
export function promptMultiLine(prompt: string): Promise<string> {
  // For simplicity, use a standard readline and let the caller handle multi-line.
  return promptUser(prompt);
}

/** Prompt with yes/no confirmation. */
export function confirmAction(message: string): Promise<boolean> {
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question(`${currentColors.warning(message + ' [y/N] ')}`, (answer) => {
      rl.close();
      resolve(answer.trim().toLowerCase() === 'y');
    });
  });
}

// ---------------------------------------------------------------------------
// Misc helpers
// ---------------------------------------------------------------------------

/** Print a horizontal rule. */
export function horizontalRule(): void {
  process.stdout.write(currentColors.muted('─'.repeat(process.stdout.columns || 80)) + '\n');
}

/** Clear the terminal screen. */
export function clearScreen(): void {
  process.stdout.write('\x1Bc');
}

/** Print the aidev ASCII banner. */
export function printBanner(): void {
  const banner = `
${currentColors.primary('   ___   _______  __    __  _______  __   __ ')}
${currentColors.primary('  |   | |   _   ||  |  |  ||   _   ||  | |  |')}
${currentColors.primary('  |   | |  |_|  ||  |  |  ||  |_|  ||  |_|  |')}
${currentColors.primary('  |   | |       ||  |__|  ||       ||       |')}
${currentColors.primary('  |   | |       ||       ||       ||       |')}
${currentColors.accent( '  |___| |_______||_______||_______||_______|')}
${currentColors.muted('  AI-powered terminal coding assistant')}
${currentColors.muted('  Supports Claude | GPT | DeepSeek | Gemini | Ollama')}
`;
  process.stdout.write(banner);
}
