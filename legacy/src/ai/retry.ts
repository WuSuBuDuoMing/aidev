/**
 * aidev CLI — Retry Logic with Exponential Backoff
 *
 * Provides resilient retry for transient AI provider errors (429, 5xx,
 * network timeouts, SSE stream errors). Inspired by MiMo-Code's unified
 * retry policy.
 *
 * @module ai/retry
 */

// ---------------------------------------------------------------------------
// Error classification
// ---------------------------------------------------------------------------

/** HTTP status codes that are safe to retry. */
const RETRYABLE_STATUS = new Set([429, 500, 502, 503, 504, 529]);

/** Network error codes that are safe to retry. */
const RETRYABLE_CODES = new Set(['ECONNRESET', 'EPIPE', 'ETIMEDOUT', 'ENETUNREACH', 'EAI_AGAIN']);

/**
 * Determine whether an error is transient and safe to retry.
 *
 * Returns `null` if the error is not retryable, or a human-readable reason
 * string if it is.
 */
export function getRetryableReason(error: unknown): string | null {
  if (!error) return null;

  const msg = error instanceof Error ? error.message : String(error);

  // Check for retryable HTTP status in error message
  const statusMatch = msg.match(/\b(429|500|502|503|504|529)\b/);
  if (statusMatch) {
    const status = parseInt(statusMatch[1], 10);
    if (RETRYABLE_STATUS.has(status)) {
      return `HTTP ${status}`;
    }
  }

  // Network errors
  const codeMatch = msg.match(/\b(ECONNRESET|EPIPE|ETIMEDOUT|ENETUNREACH|EAI_AGAIN)\b/);
  if (codeMatch && RETRYABLE_CODES.has(codeMatch[1])) {
    return `Network error: ${codeMatch[1]}`;
  }

  // SSE timeout errors
  if (msg.includes('SSE read timed out') || msg.includes('ECONNABORTED')) {
    return 'SSE stream timeout';
  }

  // Connection refused (server might be temporarily down)
  if (msg.includes('ECONNREFUSED')) {
    return 'Connection refused';
  }

  return null;
}

// ---------------------------------------------------------------------------
// Retry scheduling
// ---------------------------------------------------------------------------

export interface RetryConfig {
  /** Maximum number of retry attempts (default: 5) */
  maxRetries: number;
  /** Base delay in ms (default: 500) */
  baseDelay: number;
  /** Maximum delay cap in ms (default: 300_000 = 5 minutes) */
  maxDelay: number;
  /** Callback fired on each retry attempt */
  onRetry?: (attempt: number, delay: number, reason: string) => void;
}

const DEFAULT_RETRY_CONFIG: RetryConfig = {
  maxRetries: 5,
  baseDelay: 500,
  maxDelay: 300_000,
};

/**
 * Extract Retry-After header value in milliseconds.
 * Returns 0 if not present or unparseable.
 */
function parseRetryAfter(headers: Record<string, string | string[] | undefined>): number {
  const raw = headers['retry-after'];
  if (!raw) return 0;
  const val = Array.isArray(raw) ? raw[0] : raw;
  const seconds = parseInt(val, 10);
  if (!isNaN(seconds) && seconds > 0) return seconds * 1000;
  return 0;
}

/**
 * Calculate delay for a given attempt with exponential backoff.
 * `attempt` is 0-based. Respects Retry-After header if provided.
 */
export function calculateDelay(attempt: number, config: RetryConfig, retryAfterMs = 0): number {
  if (retryAfterMs > 0) return Math.min(retryAfterMs, config.maxDelay);
  const exponential = config.baseDelay * Math.pow(2, attempt);
  // Add jitter: ±25%
  const jitter = exponential * 0.25 * (Math.random() * 2 - 1);
  return Math.min(exponential + jitter, config.maxDelay);
}

/**
 * Execute an async function with retry logic.
 *
 * If the function throws a retryable error, it will be retried with exponential
 * backoff up to `maxRetries` times. Non-retryable errors are thrown immediately.
 */
export async function withRetry<T>(
  fn: () => Promise<T>,
  config: Partial<RetryConfig> = {},
): Promise<T> {
  const cfg = { ...DEFAULT_RETRY_CONFIG, ...config };

  for (let attempt = 0; attempt <= cfg.maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      const reason = getRetryableReason(error);

      // Not retryable — throw immediately
      if (!reason || attempt >= cfg.maxRetries) {
        throw error;
      }

      // Extract Retry-After from error if available
      const retryAfterMs = error instanceof Error && 'headers' in error
        ? parseRetryAfter((error as unknown as { headers: Record<string, string | string[] | undefined> }).headers)
        : 0;

      const delay = calculateDelay(attempt, cfg, retryAfterMs);

      if (cfg.onRetry) {
        cfg.onRetry(attempt + 1, delay, reason);
      }

      await new Promise((resolve) => setTimeout(resolve, delay));
    }
  }

  // Should not reach here, but TypeScript needs it
  throw new Error('Retry limit exceeded');
}
