/**
 * aidev CLI — Retry Logic Tests
 */

import { describe, it, expect, vi } from 'vitest';
import { getRetryableReason, calculateDelay, withRetry } from '../ai/retry';

describe('getRetryableReason', () => {
  it('returns null for non-retryable errors', () => {
    expect(getRetryableReason(null)).toBeNull();
    expect(getRetryableReason(undefined)).toBeNull();
    expect(getRetryableReason(new Error('HTTP 401: Unauthorized'))).toBeNull();
    expect(getRetryableReason(new Error('HTTP 403: Forbidden'))).toBeNull();
    expect(getRetryableReason(new Error('HTTP 400: Bad request'))).toBeNull();
    expect(getRetryableReason(new Error('HTTP 404: Not found'))).toBeNull();
  });

  it('detects retryable HTTP status codes', () => {
    expect(getRetryableReason(new Error('HTTP 429: Rate limited'))).toBe('HTTP 429');
    expect(getRetryableReason(new Error('HTTP 500: Internal error'))).toBe('HTTP 500');
    expect(getRetryableReason(new Error('API error 502'))).toBe('HTTP 502');
    expect(getRetryableReason(new Error('HTTP 503: Unavailable'))).toBe('HTTP 503');
    expect(getRetryableReason(new Error('504 Gateway Timeout'))).toBe('HTTP 504');
  });

  it('detects network errors', () => {
    expect(getRetryableReason(new Error('read ECONNRESET'))).toContain('ECONNRESET');
    expect(getRetryableReason(new Error('write EPIPE'))).toContain('EPIPE');
    expect(getRetryableReason(new Error('connect ETIMEDOUT'))).toContain('ETIMEDOUT');
  });

  it('detects SSE timeout errors', () => {
    expect(getRetryableReason(new Error('SSE read timed out'))).toContain('SSE');
    expect(getRetryableReason(new Error('ECONNABORTED'))).toContain('SSE');
  });

  it('handles string errors', () => {
    expect(getRetryableReason('HTTP 500: error')).toBe('HTTP 500');
    expect(getRetryableReason('generic error')).toBeNull();
  });
});

describe('calculateDelay', () => {
  const config = { maxRetries: 5, baseDelay: 500, maxDelay: 300_000 };

  it('returns exponential backoff on early attempts', () => {
    // Attempt 0: ~500ms (±25% jitter)
    const d0 = calculateDelay(0, config);
    expect(d0).toBeGreaterThan(300);
    expect(d0).toBeLessThan(700);

    // Attempt 1: ~1000ms
    const d1 = calculateDelay(1, config);
    expect(d1).toBeGreaterThan(700);
    expect(d1).toBeLessThan(1300);

    // Attempt 2: ~2000ms
    const d2 = calculateDelay(2, config);
    expect(d2).toBeGreaterThan(1500);
    expect(d2).toBeLessThan(2600);
  });

  it('caps at maxDelay', () => {
    const d = calculateDelay(20, config);
    expect(d).toBeLessThanOrEqual(config.maxDelay * 1.25); // include jitter
  });

  it('respects Retry-After header', () => {
    const d = calculateDelay(0, config, 5000); // 5s retry-after
    expect(d).toBe(5000);
  });

  it('caps Retry-After at maxDelay', () => {
    const d = calculateDelay(0, config, 999_999);
    expect(d).toBeLessThanOrEqual(config.maxDelay);
  });
});

describe('withRetry', () => {
  it('returns result on first success', async () => {
    const fn = vi.fn().mockResolvedValue('ok');
    const result = await withRetry(fn, { maxRetries: 3 });
    expect(result).toBe('ok');
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('retries on retryable error then succeeds', async () => {
    const fn = vi.fn()
      .mockRejectedValueOnce(new Error('HTTP 429: Rate limited'))
      .mockResolvedValue('ok');

    const result = await withRetry(fn, { maxRetries: 3, baseDelay: 10, maxDelay: 50 });
    expect(result).toBe('ok');
    expect(fn).toHaveBeenCalledTimes(2);
  });

  it('throws non-retryable errors immediately', async () => {
    const fn = vi.fn().mockRejectedValue(new Error('HTTP 401: Unauthorized'));
    await expect(withRetry(fn, { maxRetries: 3, baseDelay: 10 })).rejects.toThrow('401');
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('throws after max retries exhausted', async () => {
    const fn = vi.fn().mockRejectedValue(new Error('HTTP 500: Server error'));
    await expect(withRetry(fn, { maxRetries: 2, baseDelay: 1, maxDelay: 5 })).rejects.toThrow('500');
    expect(fn).toHaveBeenCalledTimes(3); // initial + 2 retries
  });

  it('calls onRetry callback', async () => {
    const onRetry = vi.fn();
    const fn = vi.fn()
      .mockRejectedValueOnce(new Error('ECONNRESET'))
      .mockResolvedValue('ok');

    await withRetry(fn, { maxRetries: 3, baseDelay: 1, maxDelay: 5, onRetry });
    expect(onRetry).toHaveBeenCalledTimes(1);
    expect(onRetry).toHaveBeenCalledWith(1, expect.any(Number), expect.stringContaining('ECONNRESET'));
  });
});
