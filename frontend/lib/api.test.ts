import { describe, it, expect, vi, beforeEach } from 'vitest'
import { AxiosError, AxiosHeaders } from 'axios'
import {
  getApiErrorMessage,
  isAxiosError,
  isAuthError,
  isNetworkError
} from './api'

// Helper to create mock Axios errors
function createAxiosError(
  status?: number,
  data?: { error?: string },
  code?: string
): AxiosError<{ error?: string }> {
  const error = new Error('Request failed') as AxiosError<{ error?: string }>
  error.isAxiosError = true
  error.name = 'AxiosError'
  error.code = code
  error.config = { headers: new AxiosHeaders() }
  error.toJSON = () => ({})

  if (status !== undefined) {
    error.response = {
      status,
      statusText: 'Error',
      headers: {},
      config: { headers: new AxiosHeaders() },
      data: data || {},
    }
  }

  return error
}

describe('isAxiosError', () => {
  it('returns true for Axios errors with response', () => {
    const error = createAxiosError(400, { error: 'Bad request' })
    expect(isAxiosError(error)).toBe(true)
  })

  it('returns true for Axios errors without response (network error)', () => {
    const error = createAxiosError(undefined, undefined, 'ERR_NETWORK')
    expect(isAxiosError(error)).toBe(true)
  })

  it('returns false for regular Error objects', () => {
    const error = new Error('Regular error')
    expect(isAxiosError(error)).toBe(false)
  })

  it('returns false for non-error values', () => {
    expect(isAxiosError(null)).toBe(false)
    expect(isAxiosError(undefined)).toBe(false)
    expect(isAxiosError('string error')).toBe(false)
    expect(isAxiosError({ message: 'object error' })).toBe(false)
  })
})

describe('getApiErrorMessage', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  describe('API errors with response', () => {
    it('extracts error message from API response', () => {
      const error = createAxiosError(400, { error: 'Invalid email format' })
      expect(getApiErrorMessage(error, 'Fallback')).toBe('Invalid email format')
    })

    it('handles 401 without specific message', () => {
      const error = createAxiosError(401, {})
      expect(getApiErrorMessage(error, 'Fallback')).toBe('Session expired. Please log in again.')
    })

    it('handles 403 without specific message', () => {
      const error = createAxiosError(403, {})
      expect(getApiErrorMessage(error, 'Fallback')).toBe('You do not have permission to perform this action.')
    })

    it('handles 404 without specific message', () => {
      const error = createAxiosError(404, {})
      expect(getApiErrorMessage(error, 'Fallback')).toBe('The requested resource was not found.')
    })

    it('handles 500+ server errors', () => {
      const error500 = createAxiosError(500, {})
      expect(getApiErrorMessage(error500, 'Fallback')).toBe('Server error. Please try again later.')

      const error502 = createAxiosError(502, {})
      expect(getApiErrorMessage(error502, 'Fallback')).toBe('Server error. Please try again later.')

      const error503 = createAxiosError(503, {})
      expect(getApiErrorMessage(error503, 'Fallback')).toBe('Server error. Please try again later.')
    })
  })

  describe('Network errors', () => {
    it('handles timeout error', () => {
      const error = createAxiosError(undefined, undefined, 'ECONNABORTED')
      expect(getApiErrorMessage(error, 'Fallback')).toBe('Request timed out. Please try again.')
    })

    it('handles network error', () => {
      const error = createAxiosError(undefined, undefined, 'ERR_NETWORK')
      expect(getApiErrorMessage(error, 'Fallback')).toBe('Network error. Please check your connection.')
    })

    it('handles generic connection error', () => {
      const error = createAxiosError(undefined, undefined, 'OTHER_CODE')
      expect(getApiErrorMessage(error, 'Fallback')).toBe('Unable to connect to server. Please try again.')
    })
  })

  describe('Other error types', () => {
    it('handles standard Error objects', () => {
      const error = new Error('Something went wrong')
      expect(getApiErrorMessage(error, 'Fallback')).toBe('Something went wrong')
    })

    it('handles Error objects without message', () => {
      const error = new Error('')
      expect(getApiErrorMessage(error, 'Fallback')).toBe('Fallback')
    })

    it('returns fallback for non-Error values', () => {
      expect(getApiErrorMessage(null, 'Fallback')).toBe('Fallback')
      expect(getApiErrorMessage(undefined, 'Fallback')).toBe('Fallback')
      expect(getApiErrorMessage('string', 'Fallback')).toBe('Fallback')
      expect(getApiErrorMessage(123, 'Fallback')).toBe('Fallback')
    })
  })

  describe('Logging', () => {
    it('logs error when logContext is provided', () => {
      const consoleSpy = vi.spyOn(console, 'error')
      const error = createAxiosError(400, { error: 'Test error' })

      getApiErrorMessage(error, 'Fallback', 'TestContext')

      expect(consoleSpy).toHaveBeenCalledWith('[TestContext]', error)
    })

    it('does not log when logContext is not provided', () => {
      const consoleSpy = vi.spyOn(console, 'error')
      const error = createAxiosError(400, { error: 'Test error' })

      getApiErrorMessage(error, 'Fallback')

      expect(consoleSpy).not.toHaveBeenCalled()
    })
  })
})

describe('isAuthError', () => {
  it('returns true for 401 errors', () => {
    const error = createAxiosError(401, { error: 'Unauthorized' })
    expect(isAuthError(error)).toBe(true)
  })

  it('returns true for 403 errors', () => {
    const error = createAxiosError(403, { error: 'Forbidden' })
    expect(isAuthError(error)).toBe(true)
  })

  it('returns false for other status codes', () => {
    expect(isAuthError(createAxiosError(400, {}))).toBe(false)
    expect(isAuthError(createAxiosError(404, {}))).toBe(false)
    expect(isAuthError(createAxiosError(500, {}))).toBe(false)
  })

  it('returns false for network errors', () => {
    const error = createAxiosError(undefined, undefined, 'ERR_NETWORK')
    expect(isAuthError(error)).toBe(false)
  })

  it('returns false for non-Axios errors', () => {
    expect(isAuthError(new Error('Regular error'))).toBe(false)
    expect(isAuthError(null)).toBe(false)
    expect(isAuthError(undefined)).toBe(false)
  })
})

describe('isNetworkError', () => {
  it('returns true for errors without response', () => {
    const error = createAxiosError(undefined, undefined, 'ERR_NETWORK')
    expect(isNetworkError(error)).toBe(true)
  })

  it('returns true for timeout errors', () => {
    const error = createAxiosError(undefined, undefined, 'ECONNABORTED')
    expect(isNetworkError(error)).toBe(true)
  })

  it('returns false for errors with response', () => {
    expect(isNetworkError(createAxiosError(400, {}))).toBe(false)
    expect(isNetworkError(createAxiosError(500, {}))).toBe(false)
  })

  it('returns false for non-Axios errors', () => {
    expect(isNetworkError(new Error('Regular error'))).toBe(false)
    expect(isNetworkError(null)).toBe(false)
    expect(isNetworkError(undefined)).toBe(false)
  })
})
